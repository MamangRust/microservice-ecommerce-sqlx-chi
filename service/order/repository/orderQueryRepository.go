package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/order_errors"
	"github.com/jmoiron/sqlx"
)

const getOrders = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM orders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR total_price::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrdersActive = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM orders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR total_price::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrdersTrashed = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM orders
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR total_price::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrdersByMerchant = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM orders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR total_price::TEXT ILIKE '%' || $1 || '%'
    )
    AND (
        $4::INT IS NULL
        OR merchant_id = $4
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrderByID = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at
FROM orders
WHERE
    order_id = $1
    AND deleted_at IS NULL`

type orderQueryRepository struct {
	db *sqlx.DB
}

func NewOrderQueryRepository(db *sqlx.DB) OrderQueryRepository {
	return &orderQueryRepository{
		db: db,
	}
}

func (r *orderQueryRepository) FindAll(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetOrdersRow
	err := r.db.SelectContext(ctx, &res, getOrders,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, order_errors.ErrFindAllOrders.WithInternal(err)
	}

	return res, nil
}

func (r *orderQueryRepository) FindActive(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetOrdersActiveRow
	err := r.db.SelectContext(ctx, &res, getOrdersActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, order_errors.ErrFindByActive.WithInternal(err)
	}

	return res, nil
}

func (r *orderQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetOrdersTrashedRow
	err := r.db.SelectContext(ctx, &res, getOrdersTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, order_errors.ErrFindByTrashed.WithInternal(err)
	}

	return res, nil
}

func (r *orderQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllOrderByMerchant) ([]*db.GetOrdersByMerchantRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetOrdersByMerchantRow
	err := r.db.SelectContext(ctx, &res, getOrdersByMerchant,
		req.Search,
		int32(req.PageSize),
		int32(offset),
		int32(req.MerchantID),
	)

	if err != nil {
		return nil, order_errors.ErrFindByMerchant.WithInternal(err)
	}

	return res, nil
}

func (r *orderQueryRepository) FindByID(ctx context.Context, id int) (*db.GetOrderByIDRow, error) {
	var res db.GetOrderByIDRow
	err := r.db.GetContext(ctx, &res, getOrderByID, int32(id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound.WithInternal(err)
		}
		return nil, order_errors.ErrFindById.WithInternal(err)
	}

	return &res, nil
}
