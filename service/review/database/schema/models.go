// Code migrated from sqlc-generated sources; domain models kept for repository layer.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Review struct {
	ReviewID  int32            `db:"review_id" json:"review_id"`
	UserID    int32            `db:"user_id" json:"user_id"`
	ProductID int32            `db:"product_id" json:"product_id"`
	Name      string           `db:"name" json:"name"`
	Comment   string           `db:"comment" json:"comment"`
	Rating    int32            `db:"rating" json:"rating"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
