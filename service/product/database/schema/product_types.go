// Code migrated from sqlc-generated sources; domain types kept for repository layer.
// source: products.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type AdjustProductStockParams struct {
	OperationID string `db:"operation_id" json:"operation_id"`
	ProductID   int32  `db:"product_id" json:"product_id"`
	Delta       int32  `db:"delta" json:"delta"`
}

type AdjustProductStockRow struct {
	ProductID    int32 `db:"product_id" json:"product_id"`
	CountInStock int32 `db:"count_in_stock" json:"count_in_stock"`
}

type CreateProductParams struct {
	MerchantID   int32    `db:"merchant_id" json:"merchant_id"`
	CategoryID   int32    `db:"category_id" json:"category_id"`
	Name         string   `db:"name" json:"name"`
	Description  *string  `db:"description" json:"description"`
	Price        int32    `db:"price" json:"price"`
	CountInStock int32    `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string  `db:"brand" json:"brand"`
	Weight       *int32   `db:"weight" json:"weight"`
	Rating       *float64 `db:"rating" json:"rating"`
	SlugProduct  *string  `db:"slug_product" json:"slug_product"`
	ImageProduct *string  `db:"image_product" json:"image_product"`
}

type CreateProductRow struct {
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
}

type GetProductByIDRow struct {
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
}

type GetProductsParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetProductsRow struct {
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
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetProductsActiveParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetProductsActiveRow struct {
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
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetProductsByCategoryNameParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Column2 string `db:"column2" json:"column_2"`
	Column3 int32  `db:"column3" json:"column_3"`
	Column4 int32  `db:"column4" json:"column_4"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetProductsByCategoryNameRow struct {
	TotalCount   int64            `db:"total_count" json:"total_count"`
	ProductID    int32            `db:"product_id" json:"product_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	CategoryID   int32            `db:"category_id" json:"category_id"`
	Weight       *int32           `db:"weight" json:"weight"`
	Rating       *float64         `db:"rating" json:"rating"`
	SlugProduct  *string          `db:"slug_product" json:"slug_product"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Price        int32            `db:"price" json:"price"`
	CountInStock int32            `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string          `db:"brand" json:"brand"`
	ImageProduct *string          `db:"image_product" json:"image_product"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	CategoryName interface{}      `db:"category_name" json:"category_name"`
}

type GetProductsByMerchantParams struct {
	MerchantID int32       `db:"merchant_id" json:"merchant_id"`
	Column2    *string     `db:"column2" json:"column_2"`
	Column3    interface{} `db:"column3" json:"column_3"`
	Column4    interface{} `db:"column4" json:"column_4"`
	Column5    interface{} `db:"column5" json:"column_5"`
	Limit      int32       `db:"limit" json:"limit"`
	Offset     int32       `db:"offset" json:"offset"`
}

type GetProductsByMerchantRow struct {
	TotalCount   int64            `db:"total_count" json:"total_count"`
	ProductID    int32            `db:"product_id" json:"product_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	CategoryID   int32            `db:"category_id" json:"category_id"`
	Weight       *int32           `db:"weight" json:"weight"`
	Rating       *float64         `db:"rating" json:"rating"`
	SlugProduct  *string          `db:"slug_product" json:"slug_product"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Price        int32            `db:"price" json:"price"`
	CountInStock int32            `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string          `db:"brand" json:"brand"`
	ImageProduct *string          `db:"image_product" json:"image_product"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	CategoryName string           `db:"category_name" json:"category_name"`
}

type GetProductsTrashedParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetProductsTrashedRow struct {
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
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type RestoreProductRow struct {
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
}

type TrashProductRow struct {
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
}

type UpdateProductParams struct {
	ProductID    int32    `db:"product_id" json:"product_id"`
	CategoryID   int32    `db:"category_id" json:"category_id"`
	Name         string   `db:"name" json:"name"`
	Description  *string  `db:"description" json:"description"`
	Price        int32    `db:"price" json:"price"`
	CountInStock int32    `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string  `db:"brand" json:"brand"`
	Weight       *int32   `db:"weight" json:"weight"`
	Rating       *float64 `db:"rating" json:"rating"`
	SlugProduct  *string  `db:"slug_product" json:"slug_product"`
	ImageProduct *string  `db:"image_product" json:"image_product"`
}

type UpdateProductRow struct {
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
}

type UpdateProductCountStockParams struct {
	ProductID    int32 `db:"product_id" json:"product_id"`
	CountInStock int32 `db:"count_in_stock" json:"count_in_stock"`
}

type UpdateProductCountStockRow struct {
	ProductID    int32 `db:"product_id" json:"product_id"`
	CountInStock int32 `db:"count_in_stock" json:"count_in_stock"`
}
