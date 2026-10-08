package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCategoryParams struct {
	Name          string  `db:"name" json:"name"`
	Description   *string `db:"description" json:"description"`
	SlugCategory  *string `db:"slug_category" json:"slug_category"`
	ImageCategory *string `db:"image_category" json:"image_category"`
}

type CreateCategoryRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetCategoriesRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetCategoriesActiveRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetCategoriesTrashedRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetCategoryByIDRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateCategoryParams struct {
	CategoryID    int32   `db:"category_id" json:"category_id"`
	Name          string  `db:"name" json:"name"`
	Description   *string `db:"description" json:"description"`
	SlugCategory  *string `db:"slug_category" json:"slug_category"`
	ImageCategory *string `db:"image_category" json:"image_category"`
}

type UpdateCategoryRow struct {
	CategoryID    int32            `db:"category_id" json:"category_id"`
	Name          string           `db:"name" json:"name"`
	Description   *string          `db:"description" json:"description"`
	SlugCategory  *string          `db:"slug_category" json:"slug_category"`
	ImageCategory *string          `db:"image_category" json:"image_category"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
