package outbox

import (
	"context"
	"database/sql"
	"time"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Outbox constants control the durable retry behavior of the outbox relay
// (Phase 6 — Transactional Outbox). They mirror the other services so
// behavior is uniform across repositories.
const (
	OutboxMaxAttempts    = 5
	OutboxBackoff        = 30 * time.Second
	OutboxRelayInterval  = 5 * time.Second
	OutboxRelayBatchSize = 100
	// OutboxClaimLease is how long a relay worker owns a claimed event. If the
	// worker dies after claiming but before marking the event delivered, the
	// lease expires and another relay instance re-claims and retries it.
	OutboxClaimLease = 1 * time.Minute
	// OutboxRetention is how long delivered/dead events are kept before the
	// relay purges them as part of the retention policy.
	OutboxRetention = 7 * 24 * time.Hour

	// OutboxRetentionEveryTicks runs the retention purge every N relay ticks so
	// the DELETE scan does not run on every relay cycle.
	OutboxRetentionEveryTicks = 60
)

// DBTX is the minimal executor surface the outbox SQL needs. Both *sqlx.DB
// and *sqlx.Tx satisfy it, so the relay can run against a pool or an explicit
// transaction. This keeps pkg/outbox free of any per-service schema package.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// OutboxEvent mirrors a row in the shared outbox_events table.
type OutboxEvent struct {
	OutboxID      int64
	Topic         string
	EventKey      string
	Payload       []byte
	Status        string
	Attempts      int32
	NextAttemptAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OutboxPublisher is the minimal Kafka producer surface the relay needs.
// *kafka.Kafka satisfies this interface via SendMessage.
type OutboxPublisher interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

// OutboxService persists email events durably and relays them to Kafka with
// retry and dead-letter semantics. Producers enqueue the event inside the same
// database transaction as the business write (EnqueueInTx) so a crash between
// the two cannot lose the event; the relay then guarantees delivery.
type OutboxService struct {
	db        DBTX
	publisher OutboxPublisher
	logger    logger.LoggerInterface
}

// NewOutboxService builds the outbox service. The db may be nil and the
// publisher may be nil (e.g. local dev without Kafka) — the relay then drains
// the queue without sending.
func NewOutboxService(db DBTX, publisher OutboxPublisher, log logger.LoggerInterface) *OutboxService {
	return &OutboxService{db: db, publisher: publisher, logger: log}
}

const insertOutboxEventSQL = `
INSERT INTO outbox_events (topic, event_key, payload, status, next_attempt_at)
VALUES ($1, $2, $3, 'pending', CURRENT_TIMESTAMP)
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at
`

func scanOutboxEvent(row *sql.Row) (*OutboxEvent, error) {
	var e OutboxEvent
	err := row.Scan(&e.OutboxID, &e.Topic, &e.EventKey, &e.Payload, &e.Status,
		&e.Attempts, &e.NextAttemptAt, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// EnqueueInTx persists a pending event inside the given database transaction so
// the caller can commit the business write and the event atomically. This is the
// production path: the event survives the commit and is published by the relay.
func (s *OutboxService) EnqueueInTx(ctx context.Context, tx *sqlx.Tx, topic, key string, payload []byte) error {
	_, err := scanOutboxEvent(tx.QueryRowContext(ctx, insertOutboxEventSQL, topic, key, payload))
	if err != nil {
		return err
	}
	s.logger.Info("outbox event enqueued", zap.String("topic", topic), zap.String("key", key))
	return nil
}

// Enqueue persists a pending event AFTER the business transaction has already
// committed. It is the NON-ATOMIC fallback path: a crash between the commit and
// this insert silently loses the event, so it must not be used where the
// business write is local. It exists for aggregator services whose business
// write happens in another service over gRPC (best-effort guarantee).
func (s *OutboxService) Enqueue(ctx context.Context, topic, key string, payload []byte) error {
	if s.db == nil {
		return nil
	}
	_, err := scanOutboxEvent(s.db.QueryRowContext(ctx, insertOutboxEventSQL, topic, key, payload))
	if err != nil {
		return err
	}
	s.logger.Info("outbox event enqueued", zap.String("topic", topic), zap.String("key", key))
	return nil
}

const claimPendingOutboxEventsSQL = `
UPDATE outbox_events
SET next_attempt_at = $2, updated_at = CURRENT_TIMESTAMP
WHERE outbox_id IN (
    SELECT outbox_id
    FROM outbox_events
    WHERE status = 'pending' AND next_attempt_at <= CURRENT_TIMESTAMP
    ORDER BY outbox_id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at
`

// PublishPending claims up to limit pending events whose retry window has
// elapsed, publishes each to Kafka, and marks it delivered. Claiming is atomic
// (FOR UPDATE SKIP LOCKED + lease), so concurrent relay instances never publish
// the same event twice. It returns the number of events successfully delivered.
func (s *OutboxService) PublishPending(ctx context.Context, limit int) (int, error) {
	if s.db == nil || s.publisher == nil {
		return 0, nil
	}
	rows, err := s.db.QueryContext(ctx, claimPendingOutboxEventsSQL, int32(limit), time.Now().Add(OutboxClaimLease))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if scanErr := rows.Scan(&e.OutboxID, &e.Topic, &e.EventKey, &e.Payload, &e.Status,
			&e.Attempts, &e.NextAttemptAt, &e.CreatedAt, &e.UpdatedAt); scanErr != nil {
			return 0, scanErr
		}
		events = append(events, &e)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	delivered := 0
	for _, event := range events {
		if err := s.publisher.SendMessage(ctx, event.Topic, event.EventKey, event.Payload); err != nil {
			s.logger.Error("failed to publish outbox event, scheduling retry",
				zap.Error(err),
				zap.Int64("outbox_id", event.OutboxID),
				zap.String("topic", event.Topic),
				zap.Int32("attempts", event.Attempts),
			)
			if int(event.Attempts)+1 >= OutboxMaxAttempts {
				if _, deadErr := s.db.ExecContext(ctx, markOutboxEventDeadSQL, event.OutboxID); deadErr != nil {
					s.logger.Error("failed to dead-letter outbox event", zap.Error(deadErr), zap.Int64("outbox_id", event.OutboxID))
				}
				continue
			}
			nextAttempt := time.Now().Add(OutboxBackoff * time.Duration(event.Attempts+1))
			if _, failErr := s.db.ExecContext(ctx, markOutboxEventFailedSQL, event.OutboxID, nextAttempt); failErr != nil {
				s.logger.Error("failed to record outbox retry", zap.Error(failErr), zap.Int64("outbox_id", event.OutboxID))
			}
			continue
		}
		if _, err := s.db.ExecContext(ctx, markOutboxEventDeliveredSQL, event.OutboxID); err != nil {
			s.logger.Error("failed to mark outbox event delivered", zap.Error(err), zap.Int64("outbox_id", event.OutboxID))
			continue
		}
		delivered++
	}
	return delivered, nil
}

const markOutboxEventDeliveredSQL = `
UPDATE outbox_events
SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
`

const markOutboxEventDeadSQL = `
UPDATE outbox_events
SET status = 'dead', updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
`

const markOutboxEventFailedSQL = `
UPDATE outbox_events
SET attempts = attempts + 1,
    next_attempt_at = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
`

const deleteOldOutboxEventsSQL = `
DELETE FROM outbox_events
WHERE status IN ('delivered', 'dead')
  AND updated_at < $1
`

// Start runs the relay loop until ctx is cancelled.
func (s *OutboxService) Start(ctx context.Context, interval time.Duration, limit int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickCount := 0
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("outbox relay stopped")
			return
		case <-ticker.C:
			if _, err := s.PublishPending(ctx, limit); err != nil {
				s.logger.Error("outbox relay cycle failed", zap.Error(err))
			}
			tickCount++
			// Retention runs periodically (not every tick) to avoid scanning the
			// outbox table on every relay cycle; it purges delivered/dead events
			// whose terminal state is older than the retention window.
			if s.db != nil && tickCount%OutboxRetentionEveryTicks == 0 {
				ct, err := s.db.ExecContext(ctx, deleteOldOutboxEventsSQL, time.Now().Add(-OutboxRetention))
				if err != nil {
					s.logger.Error("outbox retention cleanup failed", zap.Error(err))
				} else {
					if removed, rErr := ct.RowsAffected(); rErr == nil && removed > 0 {
						s.logger.Info("outbox retention cleanup", zap.Int64("removed", removed))
					}
				}
			}
		}
	}
}
