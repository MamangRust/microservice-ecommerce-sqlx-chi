// Code migrated from sqlc-generated sources; domain types kept for repository layer.
// source: banners.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateBannerParams struct {
	Name      string      `db:"name" json:"name"`
	StartDate pgtype.Date `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date `db:"end_date" json:"end_date"`
	StartTime pgtype.Time `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time `db:"end_time" json:"end_time"`
	IsActive  *bool       `db:"is_active" json:"is_active"`
}

type CreateBannerRow struct {
	BannerID  int32              `db:"banner_id" json:"banner_id"`
	Name      string             `db:"name" json:"name"`
	StartDate pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive  *bool              `db:"is_active" json:"is_active"`
	CreatedAt pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}

type GetBannerRow struct {
	BannerID  int32              `db:"banner_id" json:"banner_id"`
	Name      string             `db:"name" json:"name"`
	StartDate pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive  *bool              `db:"is_active" json:"is_active"`
	CreatedAt pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}

type GetBannersParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetBannersRow struct {
	BannerID   int32              `db:"banner_id" json:"banner_id"`
	Name       string             `db:"name" json:"name"`
	StartDate  pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate    pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime  pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime    pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive   *bool              `db:"is_active" json:"is_active"`
	CreatedAt  pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	TotalCount int64              `db:"total_count" json:"total_count"`
}

type GetBannersActiveParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetBannersActiveRow struct {
	BannerID   int32              `db:"banner_id" json:"banner_id"`
	Name       string             `db:"name" json:"name"`
	StartDate  pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate    pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime  pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime    pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive   *bool              `db:"is_active" json:"is_active"`
	CreatedAt  pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamptz `db:"deleted_at" json:"deleted_at"`
	TotalCount int64              `db:"total_count" json:"total_count"`
}

type GetBannersTrashedParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetBannersTrashedRow struct {
	BannerID   int32              `db:"banner_id" json:"banner_id"`
	Name       string             `db:"name" json:"name"`
	StartDate  pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate    pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime  pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime    pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive   *bool              `db:"is_active" json:"is_active"`
	CreatedAt  pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamptz `db:"deleted_at" json:"deleted_at"`
	TotalCount int64              `db:"total_count" json:"total_count"`
}

type UpdateBannerParams struct {
	BannerID  int32       `db:"banner_id" json:"banner_id"`
	Name      string      `db:"name" json:"name"`
	StartDate pgtype.Date `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date `db:"end_date" json:"end_date"`
	StartTime pgtype.Time `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time `db:"end_time" json:"end_time"`
	IsActive  *bool       `db:"is_active" json:"is_active"`
}

type UpdateBannerRow struct {
	BannerID  int32              `db:"banner_id" json:"banner_id"`
	Name      string             `db:"name" json:"name"`
	StartDate pgtype.Date        `db:"start_date" json:"start_date"`
	EndDate   pgtype.Date        `db:"end_date" json:"end_date"`
	StartTime pgtype.Time        `db:"start_time" json:"start_time"`
	EndTime   pgtype.Time        `db:"end_time" json:"end_time"`
	IsActive  *bool              `db:"is_active" json:"is_active"`
	CreatedAt pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}
