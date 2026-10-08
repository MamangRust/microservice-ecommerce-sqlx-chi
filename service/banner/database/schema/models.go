package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Banner struct {
	BannerID  int32              `db:"banner_id" json:"banner_id"`
	Name      string             `db:"name" json:"name"`
	StartDate pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive  *bool              `db:"is_active" json:"is_active"`
	CreatedAt pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamptz `db:"deleted_at" json:"deleted_at"`
}
