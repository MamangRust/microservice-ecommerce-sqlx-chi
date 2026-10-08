package inbox

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrInvalidInboxKey = errors.New("invalid consumer inbox key")

// Row is the single-scan surface the inbox SQL needs. *sql.Row satisfies it.
type Row interface {
	Scan(dest ...any) error
}

// Executor is the minimal executor surface the consumer-inbox SQL needs. Both
// *sqlx.DB and *sqlx.Tx satisfy it (schema-agnostic, no per-service schema
// package dependency).
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
}

// sqlxDB adapts *sqlx.DB to Executor: the *sql.Row it returns is a concrete
// type, while the interface exposes the Scan-only Row surface.
type sqlxDB struct{ db *sqlx.DB }

func (w sqlxDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w sqlxDB) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return w.db.QueryRowContext(ctx, query, args...)
}

// sqlxTx adapts *sqlx.Tx to Executor (see sqlxDB).
type sqlxTx struct{ tx *sqlx.Tx }

func (w sqlxTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w sqlxTx) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return w.tx.QueryRowContext(ctx, query, args...)
}

const reserveConsumerInboxSQL = `
WITH reserved AS (
    INSERT INTO consumer_inbox (
        consumer_name, event_key, topic, partition_id, message_offset,
        status, attempts, reservation_version, lease_until, last_error, processed_at
    )
    VALUES ($1, $2, $3, $4, $5, 'processing', 1, 1,
            current_timestamp + interval '1 minute', '', NULL)
    ON CONFLICT (consumer_name, event_key) DO UPDATE
    SET status = 'processing',
        attempts = consumer_inbox.attempts + 1,
        reservation_version = consumer_inbox.reservation_version + 1,
        lease_until = current_timestamp + interval '1 minute',
        last_error = '',
        topic = EXCLUDED.topic,
        partition_id = EXCLUDED.partition_id,
        message_offset = EXCLUDED.message_offset
    WHERE consumer_inbox.status <> 'processed'
      AND consumer_inbox.lease_until <= current_timestamp
    RETURNING reservation_version
)
SELECT
    EXISTS (SELECT 1 FROM reserved) AS reserved,
    EXISTS (
        SELECT 1
        FROM consumer_inbox ci
        WHERE ci.consumer_name = $1
          AND ci.event_key = $2
          AND ci.status = 'processed'
    ) AS processed,
    COALESCE(
        (SELECT reservation_version FROM reserved),
        (SELECT ci.reservation_version FROM consumer_inbox ci WHERE ci.consumer_name = $1 AND ci.event_key = $2)
    )::BIGINT AS reservation_version
`

const markConsumerInboxProcessedSQL = `
UPDATE consumer_inbox
SET status = 'processed', processed_at = current_timestamp,
    lease_until = current_timestamp, last_error = ''
WHERE consumer_name = $1 AND event_key = $2
  AND status = 'processing' AND reservation_version = $3
`

const releaseConsumerInboxSQL = `
UPDATE consumer_inbox
SET status = 'pending', lease_until = current_timestamp,
    last_error = $3
WHERE consumer_name = $1 AND event_key = $2
  AND status = 'processing' AND reservation_version = $4
`

// Reserve claims an event for a consumer. It returns false when the event was
// already processed. An expired processing lease may be reclaimed after a
// consumer crashes.
func Reserve(ctx context.Context, exec Executor, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error) {
	if exec == nil || consumerName == "" || eventKey == "" {
		return false, false, 0, ErrInvalidInboxKey
	}
	var reserved, processed bool
	var reservationVersion int64
	err := exec.QueryRowContext(ctx, reserveConsumerInboxSQL,
		consumerName, eventKey, topic, partition, offset,
	).Scan(&reserved, &processed, &reservationVersion)
	if err != nil {
		return false, false, 0, err
	}
	return reserved, processed, reservationVersion, nil
}

func MarkProcessed(ctx context.Context, exec Executor, consumerName, eventKey string, reservationVersion int64) error {
	if exec == nil || consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}
	_, err := exec.ExecContext(ctx, markConsumerInboxProcessedSQL,
		consumerName, eventKey, reservationVersion)
	return err
}

func Release(ctx context.Context, exec Executor, consumerName, eventKey string, reservationVersion int64, processingErr error) error {
	if exec == nil || consumerName == "" || eventKey == "" {
		return ErrInvalidInboxKey
	}
	lastError := "consumer processing failed"
	if processingErr != nil {
		lastError = processingErr.Error()
	}
	_, err := exec.ExecContext(ctx, releaseConsumerInboxSQL,
		consumerName, eventKey, lastError, reservationVersion)
	return err
}

// PostgresInbox adapts a *sqlx.DB to the outbox.ConsumerInbox contract.
// Reservation and completion are committed independently because an external
// side effect cannot share a PostgreSQL transaction with the Kafka consumer.
type PostgresInbox struct {
	db *sqlx.DB
}

func NewPostgresInbox(db *sqlx.DB) (*PostgresInbox, error) {
	if db == nil {
		return nil, errors.New("inbox database handle is nil")
	}
	return &PostgresInbox{db: db}, nil
}

func (i *PostgresInbox) Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error) {
	tx, err := i.db.Beginx()
	if err != nil {
		return false, false, 0, err
	}
	defer tx.Rollback()
	reserved, processed, reservationVersion, err := Reserve(ctx, sqlxTx{tx}, consumerName, eventKey, topic, partition, offset)
	if err != nil {
		return false, false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, false, 0, err
	}
	return reserved, processed, reservationVersion, nil
}

func (i *PostgresInbox) MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error {
	tx, err := i.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := MarkProcessed(ctx, sqlxTx{tx}, consumerName, eventKey, reservationVersion); err != nil {
		return err
	}
	return tx.Commit()
}

func (i *PostgresInbox) Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error {
	tx, err := i.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := Release(ctx, sqlxTx{tx}, consumerName, eventKey, reservationVersion, processingErr); err != nil {
		return err
	}
	return tx.Commit()
}
