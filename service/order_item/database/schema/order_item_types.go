// Code migrated from sqlc-generated sources; domain types kept for repository layer.
// source: order_items.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateOrderItemParams struct {
	OrderID   int32 `db:"order_id" json:"order_id"`
	ProductID int32 `db:"product_id" json:"product_id"`
	Quantity  int32 `db:"quantity" json:"quantity"`
	Price     int32 `db:"price" json:"price"`
}

type CreateOrderItemRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetOrderItemsParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetOrderItemsRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount  int64            `db:"total_count" json:"total_count"`
}

type GetOrderItemsActiveParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetOrderItemsActiveRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt   pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount  int64            `db:"total_count" json:"total_count"`
}

type GetOrderItemsByOrderRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetOrderItemsTrashedParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetOrderItemsTrashedRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt   pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount  int64            `db:"total_count" json:"total_count"`
}

type UpdateOrderItemParams struct {
	OrderItemID int32 `db:"order_item_id" json:"order_item_id"`
	Quantity    int32 `db:"quantity" json:"quantity"`
	Price       int32 `db:"price" json:"price"`
}

type UpdateOrderItemRow struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
