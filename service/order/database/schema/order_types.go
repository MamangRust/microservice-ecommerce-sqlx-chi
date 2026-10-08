// Row types migrated from sqlc-generated sources; kept for repository layer.
// source: orders.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateOrderRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetOrderByIDRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetOrdersRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetOrdersActiveRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetOrdersTrashedRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetOrdersByMerchantRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type UpdateOrderRow struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	TotalPrice int32            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
