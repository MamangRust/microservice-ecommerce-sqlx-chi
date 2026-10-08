// Hand-written models migrated from sqlc-generated source.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Product struct {
	ProductID    int32            `db:"product_id" json:"product_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	CategoryID   int32            `db:"category_id" json:"category_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Price        int32            `db:"price" json:"price"`
	CountInStock int32            `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string          `db:"brand" json:"brand"`
	Weight       *int32           `db:"weight" json:"weight"`
	Rating       *float64         `db:"rating" json:"rating"`
	SlugProduct  *string          `db:"slug_product" json:"slug_product"`
	ImageProduct *string          `db:"image_product" json:"image_product"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type ProductStockAdjustment struct {
	OperationID string           `db:"operation_id" json:"operation_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Delta       int32            `db:"delta" json:"delta"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
}
