package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/MamangRust/microservice-ecommerce-grpc-category/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/category_errors"
	"github.com/jmoiron/sqlx"
)

const createCategory = `INSERT INTO
    categories (
        name,
        description,
        slug_category,
        image_category
    )
VALUES ($1, $2, $3, $4)
RETURNING
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at`

const updateCategory = `UPDATE categories
SET
    name = $2,
    description = $3,
    slug_category = $4,
    image_category = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE
    category_id = $1
    AND deleted_at IS NULL
RETURNING
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at`

const trashCategory = `UPDATE categories
SET
    deleted_at = current_timestamp
WHERE
    category_id = $1
    AND deleted_at IS NULL
RETURNING
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at,
    deleted_at`

const restoreCategory = `UPDATE categories
SET
    deleted_at = NULL
WHERE
    category_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    category_id,
    name,
    description,
    slug_category,
    image_category,
    created_at,
    updated_at, deleted_at`

const deleteCategoryPermanently = `DELETE FROM categories
WHERE
    category_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllCategories = `UPDATE categories
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentCategories = `DELETE FROM categories WHERE deleted_at IS NOT NULL`

type categoryCommandRepository struct {
	db *sqlx.DB
}

func NewCategoryCommandRepository(db *sqlx.DB) *categoryCommandRepository {
	return &categoryCommandRepository{
		db: db,
	}
}

func (r *categoryCommandRepository) Create(ctx context.Context, request *requests.CreateCategoryRequest) (*db.CreateCategoryRow, error) {
	var category db.CreateCategoryRow
	err := r.db.GetContext(ctx, &category, createCategory,
		request.Name,
		&request.Description,
		request.SlugCategory,
		&request.ImageCategory,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound
		}
		return nil, category_errors.ErrCreateCategory.WithInternal(err)
	}

	return &category, nil
}

func (r *categoryCommandRepository) Update(ctx context.Context, request *requests.UpdateCategoryRequest) (*db.UpdateCategoryRow, error) {
	var category db.UpdateCategoryRow
	err := r.db.GetContext(ctx, &category, updateCategory,
		int32(*request.CategoryID),
		request.Name,
		&request.Description,
		request.SlugCategory,
		&request.ImageCategory,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound
		}
		return nil, category_errors.ErrUpdateCategory.WithInternal(err)
	}

	return &category, nil
}

func (r *categoryCommandRepository) Trash(ctx context.Context, category_id int) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, trashCategory, int32(category_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound
		}
		return nil, category_errors.ErrTrashedCategory.WithInternal(err)
	}

	return &category, nil
}

func (r *categoryCommandRepository) Restore(ctx context.Context, category_id int) (*db.Category, error) {
	var category db.Category
	err := r.db.GetContext(ctx, &category, restoreCategory, int32(category_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, category_errors.ErrCategoryNotFound
		}
		return nil, category_errors.ErrRestoreCategory.WithInternal(err)
	}

	return &category, nil
}

func (r *categoryCommandRepository) DeletePermanent(ctx context.Context, category_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteCategoryPermanently, int32(category_id))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete category while related products exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, category_errors.ErrCategoryNotFound
		}
		return false, category_errors.ErrDeleteCategoryPermanently.WithInternal(err)
	}

	return true, nil
}

func (r *categoryCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllCategories)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, category_errors.ErrCategoryNotFound
		}
		return false, category_errors.ErrRestoreAllCategories.WithInternal(err)
	}

	return true, nil
}

func (r *categoryCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentCategories)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete categories while related products exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, category_errors.ErrCategoryNotFound
		}
		return false, category_errors.ErrDeleteAllPermanentCategories.WithInternal(err)
	}

	return true, nil
}
