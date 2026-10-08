// Package db contains the hand-written database models and row types for the
// cart service. These structs replace the sqlc-generated models and carry
// `db` tags for sqlx struct scanning.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Cart struct {
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
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
