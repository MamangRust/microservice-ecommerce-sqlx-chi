package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order-item/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/order_item_errors"
	"github.com/jmoiron/sqlx"
)

const getOrderItems = `SELECT
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM order_items
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR product_id::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrderItemsActive = `SELECT
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM order_items
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR product_id::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getOrderItemsTrashed = `SELECT
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM order_items
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR product_id::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY deleted_at DESC
LIMIT $2
OFFSET
    $3`

const getOrderItemsByOrder = `SELECT
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at
FROM order_items
WHERE
    order_id = $1
    AND deleted_at IS NULL`

type orderItemQueryRepository struct {
	db *sqlx.DB
}

func NewOrderItemQueryRepository(db *sqlx.DB) *orderItemQueryRepository {
	return &orderItemQueryRepository{
		db: db,
	}
}

func (r *orderItemQueryRepository) FindAll(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var items []*db.GetOrderItemsRow
	err := r.db.SelectContext(ctx, &items, getOrderItems,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, orderitem_errors.ErrFindAllOrderItems
	}

	return items, nil
}

func (r *orderItemQueryRepository) FindActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var items []*db.GetOrderItemsActiveRow
	err := r.db.SelectContext(ctx, &items, getOrderItemsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, orderitem_errors.ErrFindByActive
	}

	return items, nil
}

func (r *orderItemQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var items []*db.GetOrderItemsTrashedRow
	err := r.db.SelectContext(ctx, &items, getOrderItemsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, orderitem_errors.ErrFindByTrashed
	}

	return items, nil
}

func (r *orderItemQueryRepository) FindOrderItemByOrder(ctx context.Context, order_id int) ([]*db.GetOrderItemsByOrderRow, error) {
	var items []*db.GetOrderItemsByOrderRow
	err := r.db.SelectContext(ctx, &items, getOrderItemsByOrder, int32(order_id))

	if err != nil {
		return nil, orderitem_errors.ErrFindOrderItemByOrder
	}

	return items, nil
}
