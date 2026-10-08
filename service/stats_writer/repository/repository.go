package repository

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-shared/domain/events"
)

// Repository materializes stats events into ClickHouse. Batches are flushed on
// size and on an interval so a burst of events does not hammer ClickHouse with
// single-row inserts.
//
// eventID is the idempotency key stored as part of the ClickHouse primary key;
// eventVersion feeds the ReplacingMergeTree version column so a newer delivery
// (or a re-backfill with a newer run timestamp) supersedes older rows.
type Repository interface {
	InsertOrderEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderEvent) error
	InsertOrderItemEvent(ctx context.Context, eventID string, eventVersion uint64, event events.OrderItemEvent) error
	InsertTransactionEvent(ctx context.Context, eventID string, eventVersion uint64, event events.TransactionEvent) error

	Flush(ctx context.Context) error
	Close() error
}
