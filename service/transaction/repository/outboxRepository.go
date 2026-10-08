package repository

import (
	"context"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-grpc-transaction/database/schema"
	"github.com/jmoiron/sqlx"
)

const createOutboxEvent = `INSERT INTO outbox_events (topic, event_key, payload, status, next_attempt_at)
VALUES ($1, $2, $3, 'pending', CURRENT_TIMESTAMP)
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at`

const getPendingOutboxEvents = `SELECT outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at
FROM outbox_events
WHERE status = 'pending' AND next_attempt_at <= CURRENT_TIMESTAMP
ORDER BY outbox_id
LIMIT $1`

const claimPendingOutboxEvents = `UPDATE outbox_events
SET next_attempt_at = $2, updated_at = CURRENT_TIMESTAMP
WHERE outbox_id IN (
    SELECT outbox_id
    FROM outbox_events
    WHERE status = 'pending' AND next_attempt_at <= CURRENT_TIMESTAMP
    ORDER BY outbox_id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at`

const markOutboxEventDelivered = `UPDATE outbox_events
SET status = 'delivered', updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at`

const markOutboxEventFailed = `UPDATE outbox_events
SET attempts = attempts + 1,
    next_attempt_at = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at`

const markOutboxEventDead = `UPDATE outbox_events
SET status = 'dead', updated_at = CURRENT_TIMESTAMP
WHERE outbox_id = $1 AND status = 'pending'
RETURNING outbox_id, topic, event_key, payload, status, attempts, next_attempt_at, created_at, updated_at`

const deleteOldOutboxEvents = `DELETE FROM outbox_events
WHERE status IN ('delivered', 'dead')
  AND updated_at < $1`

// OutboxRepository provides durable persistence for events that must be published
// to Kafka after the business transaction commits. A relay service consumes
// pending events and retries until delivered or dead-lettered.
type OutboxRepository interface {
	Create(ctx context.Context, topic, key string, payload []byte) (*db.OutboxEvent, error)

	// CreateInTx persists a pending event inside the given database transaction so
	// the caller can commit the business write and the event atomically.
	CreateInTx(ctx context.Context, tx *sqlx.Tx, topic, key string, payload []byte) (*db.OutboxEvent, error)

	GetPending(ctx context.Context, limit int) ([]*db.OutboxEvent, error)

	// Claim atomically claims up to limit pending events whose retry window has
	// elapsed by extending next_attempt_at to leaseUntil. The FOR UPDATE SKIP
	// LOCKED guard ensures concurrent relay instances never publish the same
	// event; a crashed worker's claim simply expires when the lease passes.
	Claim(ctx context.Context, limit int, leaseUntil time.Time) ([]*db.OutboxEvent, error)

	MarkDelivered(ctx context.Context, outboxID int64) (*db.OutboxEvent, error)
	MarkFailed(ctx context.Context, outboxID int64, nextAttemptAt time.Time) (*db.OutboxEvent, error)
	MarkDead(ctx context.Context, outboxID int64) (*db.OutboxEvent, error)
	DeleteOld(ctx context.Context, cutoff time.Time) (int64, error)
}

type outboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(queries *sqlx.DB) OutboxRepository {
	return &outboxRepository{db: queries}
}

func (r *outboxRepository) Create(ctx context.Context, topic, key string, payload []byte) (*db.OutboxEvent, error) {
	var event db.OutboxEvent

	if err := r.db.GetContext(ctx, &event, createOutboxEvent, topic, key, payload); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *outboxRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, topic, key string, payload []byte) (*db.OutboxEvent, error) {
	var event db.OutboxEvent

	if err := tx.GetContext(ctx, &event, createOutboxEvent, topic, key, payload); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *outboxRepository) GetPending(ctx context.Context, limit int) ([]*db.OutboxEvent, error) {
	var events []*db.OutboxEvent

	if err := r.db.SelectContext(ctx, &events, getPendingOutboxEvents, int32(limit)); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) Claim(ctx context.Context, limit int, leaseUntil time.Time) ([]*db.OutboxEvent, error) {
	var events []*db.OutboxEvent

	if err := r.db.SelectContext(ctx, &events, claimPendingOutboxEvents, int32(limit), leaseUntil); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) MarkDelivered(ctx context.Context, outboxID int64) (*db.OutboxEvent, error) {
	var event db.OutboxEvent

	if err := r.db.GetContext(ctx, &event, markOutboxEventDelivered, outboxID); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *outboxRepository) MarkFailed(ctx context.Context, outboxID int64, nextAttemptAt time.Time) (*db.OutboxEvent, error) {
	var event db.OutboxEvent

	if err := r.db.GetContext(ctx, &event, markOutboxEventFailed, outboxID, nextAttemptAt); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *outboxRepository) MarkDead(ctx context.Context, outboxID int64) (*db.OutboxEvent, error) {
	var event db.OutboxEvent

	if err := r.db.GetContext(ctx, &event, markOutboxEventDead, outboxID); err != nil {
		return nil, err
	}

	return &event, nil
}

func (r *outboxRepository) DeleteOld(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, deleteOldOutboxEvents, cutoff)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}
