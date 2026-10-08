// Code migrated from sqlc-generated sources; domain types kept for repository layer.
// source: sliders.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateSliderParams struct {
	Name  string `db:"name" json:"name"`
	Image string `db:"image" json:"image"`
}

type CreateSliderRow struct {
	SliderID  int32            `db:"slider_id" json:"slider_id"`
	Name      string           `db:"name" json:"name"`
	Image     string           `db:"image" json:"image"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetSliderByIDRow struct {
	SliderID  int32            `db:"slider_id" json:"slider_id"`
	Name      string           `db:"name" json:"name"`
	Image     string           `db:"image" json:"image"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetSlidersParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetSlidersRow struct {
	SliderID   int32            `db:"slider_id" json:"slider_id"`
	Name       string           `db:"name" json:"name"`
	Image      string           `db:"image" json:"image"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetSlidersActiveParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetSlidersActiveRow struct {
	SliderID   int32            `db:"slider_id" json:"slider_id"`
	Name       string           `db:"name" json:"name"`
	Image      string           `db:"image" json:"image"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetSlidersTrashedParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetSlidersTrashedRow struct {
	SliderID   int32            `db:"slider_id" json:"slider_id"`
	Name       string           `db:"name" json:"name"`
	Image      string           `db:"image" json:"image"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type UpdateSliderParams struct {
	SliderID int32  `db:"slider_id" json:"slider_id"`
	Name     string `db:"name" json:"name"`
	Image    string `db:"image" json:"image"`
}

type UpdateSliderRow struct {
	SliderID  int32            `db:"slider_id" json:"slider_id"`
	Name      string           `db:"name" json:"name"`
	Image     string           `db:"image" json:"image"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
