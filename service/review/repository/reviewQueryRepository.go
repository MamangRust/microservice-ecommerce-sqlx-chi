package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-review/database/schema"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	review_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/review"
	"github.com/jmoiron/sqlx"
)

const getReviews = `SELECT
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM reviews
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR review_id::TEXT ILIKE '%' || $1 || '%'
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getReviewsActive = `SELECT
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM reviews
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR review_id::TEXT ILIKE '%' || $1 || '%'
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getReviewsTrashed = `SELECT
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM reviews
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR review_id::TEXT ILIKE '%' || $1 || '%'
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getReviewByProductId = `SELECT
    r.review_id,
    r.user_id,
    r.product_id,
    r.name,
    r.comment,
    r.rating,
    r.created_at,
    r.updated_at,
    r.deleted_at,
    COUNT(*) OVER () AS total_count,
    '[]'::jsonb AS review_details
FROM reviews r
WHERE
    r.deleted_at IS NULL
    AND r.product_id = $1
    AND (
        $2::INT IS NULL
        OR r.rating = $2
    )
ORDER BY r.created_at DESC
LIMIT $3
OFFSET
    $4`

const getReviewByMerchantId = `SELECT
    r.review_id,
    r.user_id,
    r.product_id,
    r.name,
    r.comment,
    r.rating,
    r.created_at,
    r.updated_at,
    r.deleted_at,
    COUNT(*) OVER () AS total_count,
    '[]'::jsonb AS review_details
FROM reviews r
WHERE
    r.deleted_at IS NULL
    AND r.product_id = ANY($1::int[])
    AND (
        $2::INT IS NULL
        OR r.rating = $2
    )
ORDER BY r.created_at DESC
LIMIT $3
OFFSET
    $4`

const getReviewByID = `SELECT
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at
FROM reviews
WHERE
    review_id = $1
    AND deleted_at IS NULL`

type reviewQueryRepository struct {
	db          *sqlx.DB
	productRepo productadapter.QueryRepository
}

func NewReviewQueryRepository(db *sqlx.DB, productRepo productadapter.QueryRepository) *reviewQueryRepository {
	return &reviewQueryRepository{
		db:          db,
		productRepo: productRepo,
	}
}

func (r *reviewQueryRepository) FindAll(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var reviews []*db.GetReviewsRow
	err := r.db.SelectContext(ctx, &reviews, getReviews,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_errors.ErrFindAllReviews.WithInternal(err)
	}

	return reviews, nil
}

func (r *reviewQueryRepository) FindByProduct(ctx context.Context, req *requests.FindAllReviewByProduct) ([]*db.GetReviewByProductIdRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var reviews []*db.GetReviewByProductIdRow
	err := r.db.SelectContext(ctx, &reviews, getReviewByProductId,
		int32(req.ProductID),
		int32(req.Rating),
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_errors.ErrFindReviewsByProduct.WithInternal(err)
	}

	return reviews, nil
}

func (r *reviewQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllReviewByMerchant) ([]*db.GetReviewByMerchantIdRow, error) {
	// The review service owns no products table (per-service DB split), so the
	// merchant's product IDs are resolved via the product service gRPC before
	// querying reviews. Fetch a large page to collect every product ID.
	productIDs, err := r.productRepo.FindIDsByMerchant(ctx, req.MerchantID)
	if err != nil {
		return nil, review_errors.ErrFindReviewsByMerchant.WithInternal(err)
	}

	offset := (req.Page - 1) * req.PageSize

	var reviews []*db.GetReviewByMerchantIdRow
	err = r.db.SelectContext(ctx, &reviews, getReviewByMerchantId,
		productIDs,
		int32(req.Rating),
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_errors.ErrFindReviewsByMerchant.WithInternal(err)
	}

	return reviews, nil
}

func (r *reviewQueryRepository) FindActive(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var reviews []*db.GetReviewsActiveRow
	err := r.db.SelectContext(ctx, &reviews, getReviewsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_errors.ErrFindActiveReviews.WithInternal(err)
	}

	return reviews, nil
}

func (r *reviewQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var reviews []*db.GetReviewsTrashedRow
	err := r.db.SelectContext(ctx, &reviews, getReviewsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, review_errors.ErrFindTrashedReviews.WithInternal(err)
	}

	return reviews, nil
}

func (r *reviewQueryRepository) FindByID(ctx context.Context, id int) (*db.GetReviewByIDRow, error) {
	var review db.GetReviewByIDRow
	err := r.db.GetContext(ctx, &review, getReviewByID, int32(id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, review_errors.ErrReviewNotFound
		}
		return nil, review_errors.ErrFindReviewByID.WithInternal(err)
	}

	return &review, nil
}
