package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-review/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	review_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/review"
	"github.com/jmoiron/sqlx"
)

const createReview = `INSERT INTO
    reviews (
        user_id,
        product_id,
        name,
        comment,
        rating
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at`

const updateReview = `UPDATE reviews
SET
    name = $2,
    comment = $3,
    rating = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE
    review_id = $1
    AND deleted_at IS NULL
RETURNING
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at`

const trashReview = `UPDATE reviews
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    review_id = $1
    AND deleted_at IS NULL
RETURNING
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    deleted_at`

const restoreReview = `UPDATE reviews
SET
    deleted_at = NULL
WHERE
    review_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    deleted_at`

const deleteReviewPermanently = `DELETE FROM reviews WHERE review_id = $1 AND deleted_at IS NOT NULL`

const restoreAllReviews = `UPDATE reviews SET deleted_at = NULL WHERE deleted_at IS NOT NULL`

const deleteAllPermanentReviews = `DELETE FROM reviews WHERE deleted_at IS NOT NULL`

type reviewCommandRepository struct {
	db *sqlx.DB
}

func NewReviewCommandRepository(db *sqlx.DB) *reviewCommandRepository {
	return &reviewCommandRepository{
		db: db,
	}
}

func (r *reviewCommandRepository) Create(ctx context.Context, request *requests.CreateReviewRequest) (*db.CreateReviewRow, error) {
	// CreateReviewRequest carries no display name, so the name column keeps
	// the empty value it received before the sqlx migration.
	var review db.CreateReviewRow
	err := r.db.GetContext(ctx, &review, createReview,
		int32(request.UserID),
		int32(request.ProductID),
		"",
		request.Comment,
		int32(request.Rating),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_errors.ErrReviewNotFound
		}
		return nil, review_errors.ErrCreateReview.WithInternal(err)
	}

	return &review, nil
}

func (r *reviewCommandRepository) Update(ctx context.Context, request *requests.UpdateReviewRequest) (*db.UpdateReviewRow, error) {
	var review db.UpdateReviewRow
	err := r.db.GetContext(ctx, &review, updateReview,
		int32(*request.ReviewID),
		request.Name,
		request.Comment,
		int32(request.Rating),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_errors.ErrReviewNotFound
		}
		return nil, review_errors.ErrUpdateReview.WithInternal(err)
	}

	return &review, nil
}

func (r *reviewCommandRepository) Trash(ctx context.Context, review_id int) (*db.Review, error) {
	var review db.Review
	err := r.db.GetContext(ctx, &review, trashReview, int32(review_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_errors.ErrReviewNotFound
		}
		return nil, review_errors.ErrTrashReview.WithInternal(err)
	}

	return &review, nil
}

func (r *reviewCommandRepository) Restore(ctx context.Context, review_id int) (*db.Review, error) {
	var review db.Review
	err := r.db.GetContext(ctx, &review, restoreReview, int32(review_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_errors.ErrReviewNotFound
		}
		return nil, review_errors.ErrRestoreReview.WithInternal(err)
	}

	return &review, nil
}

func (r *reviewCommandRepository) DeletePermanent(ctx context.Context, review_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteReviewPermanently, int32(review_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_errors.ErrReviewNotFound
		}
		return false, review_errors.ErrDeleteReviewPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *reviewCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllReviews); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_errors.ErrReviewNotFound
		}
		return false, review_errors.ErrRestoreAllReviews.WithInternal(err)
	}
	return true, nil
}

func (r *reviewCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentReviews); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_errors.ErrReviewNotFound
		}
		return false, review_errors.ErrDeleteAllPermanentReview.WithInternal(err)
	}
	return true, nil
}
