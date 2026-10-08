package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-category/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/category_errors"
	"github.com/jmoiron/sqlx"
)

const getCategories = `SELECT
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM categories
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR slug_category ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getCategoriesActive = `SELECT
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM categories
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR slug_category ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getCategoriesTrashed = `SELECT
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM categories
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR slug_category ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getCategoryByID = `SELECT
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at
FROM categories
WHERE
    category_id = $1
    AND deleted_at IS NULL`

const getCategoryByIDTrashed = `SELECT category_id, name, description, slug_category, image_category, created_at, updated_at, deleted_at
FROM categories
WHERE
    category_id = $1
    AND deleted_at IS NOT NULL`

type categoryQueryRepository struct {
	db *sqlx.DB
}

func NewCategoryQueryRepository(db *sqlx.DB) *categoryQueryRepository {
	return &categoryQueryRepository{
		db: db,
	}
}

func (r *categoryQueryRepository) FindAll(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetCategoriesRow
	err := r.db.SelectContext(ctx, &res, getCategories,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, category_errors.ErrFindAllCategory.WithInternal(err)
	}

	return res, nil
}

func (r *categoryQueryRepository) FindActive(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetCategoriesActiveRow
	err := r.db.SelectContext(ctx, &res, getCategoriesActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, category_errors.ErrFindByActiveCategory.WithInternal(err)
	}

	return res, nil
}

func (r *categoryQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetCategoriesTrashedRow
	err := r.db.SelectContext(ctx, &res, getCategoriesTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, category_errors.ErrFindByTrashedCategory.WithInternal(err)
	}

	return res, nil
}

func (r *categoryQueryRepository) FindByID(ctx context.Context, category_id int) (*db.GetCategoryByIDRow, error) {
	var category db.GetCategoryByIDRow
	err := r.db.GetContext(ctx, &category, getCategoryByID, int32(category_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound.WithInternal(err)
		}
		return nil, category_errors.ErrFindCategoryById.WithInternal(err)
	}

	return &category, nil
}

func (r *categoryQueryRepository) FindByIDTrashed(ctx context.Context, category_id int) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, getCategoryByIDTrashed, int32(category_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound.WithInternal(err)
		}
		return nil, category_errors.ErrFindCategoryByIdTrashed.WithInternal(err)
	}

	return &category, nil
}
