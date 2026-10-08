package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Category struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
