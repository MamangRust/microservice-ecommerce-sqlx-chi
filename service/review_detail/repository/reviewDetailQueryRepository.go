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

const getReviewDetails = `SELECT
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM review_details
WHERE
    LOWER(COALESCE(caption, '')) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getReviewDetailsActive = `SELECT
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM review_details
WHERE
    deleted_at IS NULL
    AND LOWER(COALESCE(caption, '')) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getReviewDetailsTrashed = `SELECT
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM review_details
WHERE
    deleted_at IS NOT NULL
    AND LOWER(COALESCE(caption, '')) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getReviewDetail = `SELECT
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at
FROM review_details
WHERE
    review_detail_id = $1
    AND deleted_at IS NULL;`

const getReviewDetailTrashed = `SELECT *
FROM review_details
WHERE
    review_detail_id = $1
    AND deleted_at IS NOT NULL;`

type reviewDetailQueryRepository struct {
	db *sqlx.DB
}

func NewReviewDetailQueryRepository(db *sqlx.DB) *reviewDetailQueryRepository {
	return &reviewDetailQueryRepository{
		db: db,
	}
}

func (r *reviewDetailQueryRepository) FindAll(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewDetailsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetReviewDetailsRow
	err := r.db.SelectContext(ctx, &res, getReviewDetails,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_detail_errors.ErrFindAllReviewDetails.WithInternal(err)
	}

	return res, nil
}

func (r *reviewDetailQueryRepository) FindActive(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewDetailsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetReviewDetailsActiveRow
	err := r.db.SelectContext(ctx, &res, getReviewDetailsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_detail_errors.ErrFindActiveReviewDetails.WithInternal(err)
	}

	return res, nil
}

func (r *reviewDetailQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewDetailsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetReviewDetailsTrashedRow
	err := r.db.SelectContext(ctx, &res, getReviewDetailsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_detail_errors.ErrFindTrashedReviewDetails.WithInternal(err)
	}

	return res, nil
}

func (r *reviewDetailQueryRepository) FindByID(ctx context.Context, user_id int) (*db.GetReviewDetailRow, error) {
	var res db.GetReviewDetailRow
	err := r.db.GetContext(ctx, &res, getReviewDetail, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrFindByIdReviewDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *reviewDetailQueryRepository) FindByIDTrashed(ctx context.Context, user_id int) (*db.ReviewDetail, error) {
	var res db.ReviewDetail
	err := r.db.GetContext(ctx, &res, getReviewDetailTrashed, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_detail_errors.ErrReviewDetailNotFound
		}
		return nil, review_detail_errors.ErrFindByIdTrashedReviewDetail.WithInternal(err)
	}

	return &res, nil
}
