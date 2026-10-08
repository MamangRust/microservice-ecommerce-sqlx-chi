package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
	"github.com/jmoiron/sqlx"
)

type merchantQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantQueryRepository(db *sqlx.DB) MerchantQueryRepository {
	return &merchantQueryRepository{
		db: db,
	}
}

const getMerchants = `SELECT
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR contact_email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getMerchantsActive = `SELECT
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR contact_email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getMerchantsTrashed = `SELECT
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR contact_email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

const getMerchantByID = `SELECT
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at
FROM merchants
WHERE
    merchant_id = $1
    AND deleted_at IS NULL`

func (r *merchantQueryRepository) FindAll(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetMerchantsRow
	err := r.db.SelectContext(ctx, &merchants, getMerchants,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, merchant_errors.ErrFindAllMerchants.WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindActive(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetMerchantsActiveRow
	err := r.db.SelectContext(ctx, &merchants, getMerchantsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, merchant_errors.ErrFindActiveMerchants.WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var merchants []*db.GetMerchantsTrashedRow
	err := r.db.SelectContext(ctx, &merchants, getMerchantsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, merchant_errors.ErrFindTrashedMerchants.WithInternal(err)
	}

	return merchants, nil
}

func (r *merchantQueryRepository) FindByID(ctx context.Context, user_id int) (*db.GetMerchantByIDRow, error) {
	var merchant db.GetMerchantByIDRow
	err := r.db.GetContext(ctx, &merchant, getMerchantByID, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound.WithInternal(err)
		}
		return nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	return &merchant, nil
}
