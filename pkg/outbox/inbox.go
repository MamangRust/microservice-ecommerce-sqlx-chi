package outbox

import (
	"context"
)

// ConsumerInbox is the durable deduplication contract used by Kafka handlers
// (Phase 3 — Durable Idempotency). It replaces in-memory-only deduplication:
// reservations survive consumer restarts and rebalances, so at-least-once
// redelivery cannot send the same email twice.
//
// The PostgreSQL-backed implementation lives in the email service
// (internal/inbox), which owns the consumer_inbox table and its generated
// queries; this package only defines the contract so handlers stay decoupled
// from the concrete store.
type ConsumerInbox interface {
	Reserve(ctx context.Context, consumerName, eventKey, topic string, partition int32, offset int64) (bool, bool, int64, error)
	MarkProcessed(ctx context.Context, consumerName, eventKey string, reservationVersion int64) error
	Release(ctx context.Context, consumerName, eventKey string, reservationVersion int64, processingErr error) error
}
