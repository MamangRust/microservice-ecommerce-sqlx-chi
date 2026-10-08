// Package db contains the hand-written database models for the transaction
// service. These structs replace the sqlc-generated models and carry `db`
// tags for sqlx struct scanning.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxEvent struct {
	OutboxID      int64            `db:"outbox_id" json:"outbox_id"`
	Topic         string           `db:"topic" json:"topic"`
	EventKey      string           `db:"event_key" json:"event_key"`
	Payload       []byte           `db:"payload" json:"payload"`
	Status        string           `db:"status" json:"status"`
	Attempts      int32            `db:"attempts" json:"attempts"`
	NextAttemptAt time.Time        `db:"next_attempt_at" json:"next_attempt_at"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type Transaction struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
