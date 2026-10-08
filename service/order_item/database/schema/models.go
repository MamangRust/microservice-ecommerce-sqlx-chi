// Hand-written models migrated from sqlc-generated source.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderItem struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt   pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
