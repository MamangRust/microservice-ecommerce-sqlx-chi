package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCartRow struct {
	CartID    int32            `db:"cart_id" json:"cart_id"`
	UserID    int32            `db:"user_id" json:"user_id"`
	ProductID int32            `db:"product_id" json:"product_id"`
	Name      string           `db:"name" json:"name"`
	Price     int32            `db:"price" json:"price"`
	Image     string           `db:"image" json:"image"`
	Quantity  int32            `db:"quantity" json:"quantity"`
	Weight    int32            `db:"weight" json:"weight"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetCartsRow struct {
	CartID     int32            `db:"cart_id" json:"cart_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	ProductID  int32            `db:"product_id" json:"product_id"`
	Name       string           `db:"name" json:"name"`
	Price      int32            `db:"price" json:"price"`
	Image      string           `db:"image" json:"image"`
	Quantity   int32            `db:"quantity" json:"quantity"`
	Weight     int32            `db:"weight" json:"weight"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}
