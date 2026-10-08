// Hand-written models migrated from sqlc-generated source.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type OrderStockReservation struct {
	ReservationID int32            `db:"reservation_id" json:"reservation_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	ProductID     int32            `db:"product_id" json:"product_id"`
	Quantity      int32            `db:"quantity" json:"quantity"`
	Status        string           `db:"status" json:"status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

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

type ProductStockAdjustment struct {
	OperationID string           `db:"operation_id" json:"operation_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Delta       int32            `db:"delta" json:"delta"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
}
