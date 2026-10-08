package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	review_detail_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/review_detail"
)

const createReviewDetail = `INSERT INTO
    review_details (review_id, type, url, caption)
VALUES ($1, $2, $3, $4)
RETURNING
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at;`

const updateReviewDetail = `UPDATE review_details
SET
    type = $1,
    url = $2,
    caption = $3
WHERE
    review_detail_id = $4
RETURNING
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at;`

const trashReviewDetail = `UPDATE review_details
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    review_detail_id = $1
    AND deleted_at IS NULL
RETURNING
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at,
    deleted_at;`

const restoreReviewDetail = `UPDATE review_details
SET
    deleted_at = NULL
WHERE
    review_detail_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at,
    deleted_at;`

const deletePermanentReviewDetail = `DELETE FROM review_details
WHERE
    review_detail_id = $1
    AND deleted_at IS NOT NULL;`

const restoreAllReviewDetails = `UPDATE review_details
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL;`

const deleteAllPermanentReviewDetails = `DELETE FROM review_details WHERE deleted_at IS NOT NULL;`

type reviewDetailCommandRepository struct {
	db *sqlx.DB
}

func NewReviewDetailCommandRepository(db *sqlx.DB) *reviewDetailCommandRepository {
	return &reviewDetailCommandRepository{
		db: db,
	}
}

func (r *reviewDetailCommandRepository) Create(ctx context.Context, request *requests.CreateReviewDetailRequest) (*db.CreateReviewDetailRow, error) {
	var reviewDetail db.CreateReviewDetailRow
	err := r.db.GetContext(ctx, &reviewDetail, createReviewDetail,
		int32(request.ReviewID),
		request.Type,
		request.Url,
		stringPtr(request.Caption),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrCreateReviewDetail.WithInternal(err)
	}

	return &reviewDetail, nil
}

func (r *reviewDetailCommandRepository) Update(ctx context.Context, request *requests.UpdateReviewDetailRequest) (*db.UpdateReviewDetailRow, error) {
	var res db.UpdateReviewDetailRow
	err := r.db.GetContext(ctx, &res, updateReviewDetail,
		request.Type,
		request.Url,
		stringPtr(request.Caption),
		int32(*request.ReviewDetailID),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrUpdateReviewDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *reviewDetailCommandRepository) Trash(ctx context.Context, ReviewDetail_id int) (*db.ReviewDetail, error) {
	var res db.ReviewDetail
	err := r.db.GetContext(ctx, &res, trashReviewDetail, int32(ReviewDetail_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrTrashedReviewDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *reviewDetailCommandRepository) Restore(ctx context.Context, ReviewDetail_id int) (*db.ReviewDetail, error) {
	var res db.ReviewDetail
	err := r.db.GetContext(ctx, &res, restoreReviewDetail, int32(ReviewDetail_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrRestoreReviewDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *reviewDetailCommandRepository) DeletePermanent(ctx context.Context, ReviewDetail_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deletePermanentReviewDetail, int32(ReviewDetail_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_detail_errors.ErrReviewDetailNotFound
		}
		return false, review_detail_errors.ErrDeleteReviewDetailPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *reviewDetailCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllReviewDetails)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_detail_errors.ErrReviewDetailNotFound
		}
		return false, review_detail_errors.ErrRestoreAllReviewDetails.WithInternal(err)
	}
	return true, nil
}

func (r *reviewDetailCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentReviewDetails)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, review_detail_errors.ErrReviewDetailNotFound
		}
		return false, review_detail_errors.ErrDeleteAllReviewDetails.WithInternal(err)
	}
	return true, nil
}

func stringPtr(s string) *string {
	return &s
}
