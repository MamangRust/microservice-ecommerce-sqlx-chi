// Code migrated from sqlc-generated models; column mapping kept for sqlx.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Slider struct {
	SliderID  int32            `db:"slider_id" json:"slider_id"`
	Name      string           `db:"name" json:"name"`
	Image     string           `db:"image" json:"image"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
