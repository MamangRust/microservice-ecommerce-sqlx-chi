// Code migrated from sqlc-generated sources; domain models kept for repository layer.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ReviewDetail struct {
	ReviewDetailID int32            `db:"review_detail_id" json:"review_detail_id"`
	ReviewID       int32            `db:"review_id" json:"review_id"`
	Type           string           `db:"type" json:"type"`
	Url            string           `db:"url" json:"url"`
	Caption        *string          `db:"caption" json:"caption"`
	CreatedAt      pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt      pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt      pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
