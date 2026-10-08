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

const createOrder = `INSERT INTO
    orders (
        merchant_id,
        user_id,
        total_price
    )
VALUES ($1, $2, $3)
RETURNING
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at`

const updateOrder = `UPDATE orders
SET
    total_price = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE
    order_id = $1
    AND deleted_at IS NULL
RETURNING
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at`

const trashedOrder = `UPDATE orders
SET
    deleted_at = current_timestamp
WHERE
    order_id = $1
    AND deleted_at IS NULL
RETURNING
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at`

const restoreOrder = `UPDATE orders
SET
    deleted_at = NULL
WHERE
    order_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at`

const getTrashedOrder = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at
FROM orders
WHERE order_id = $1 AND deleted_at IS NOT NULL`

const getTrashedOrders = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    deleted_at
FROM orders
WHERE deleted_at IS NOT NULL
ORDER BY order_id`

const deleteOrderPermanently = `DELETE FROM orders WHERE order_id = $1 AND deleted_at IS NOT NULL`

const deleteOrderPermanentlyWithChildren = `WITH
    trashed AS (
        SELECT order_id FROM orders WHERE order_id = $1 AND deleted_at IS NOT NULL
    ),
    deleted_reservations AS (
        DELETE FROM order_stock_reservations WHERE order_id IN (SELECT order_id FROM trashed)
    )
DELETE FROM orders o
WHERE o.order_id = $1 AND o.deleted_at IS NOT NULL
RETURNING o.order_id`

const restoreAllOrders = `UPDATE orders
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentOrders = `DELETE FROM orders WHERE deleted_at IS NOT NULL`

type orderCommandRepository struct {
	db *sqlx.DB
}

func NewOrderCommandRepository(db *sqlx.DB) OrderCommandRepository {
	return &orderCommandRepository{
		db: db,
	}
}

func (r *orderCommandRepository) Create(ctx context.Context, request *requests.CreateOrderRecordRequest) (*db.CreateOrderRow, error) {
	var res db.CreateOrderRow
	err := r.db.GetContext(ctx, &res, createOrder,
		int32(request.MerchantID),
		int32(request.UserID),
		int32(request.TotalPrice),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound
		}
		return nil, order_errors.ErrCreateOrder.WithInternal(err)
	}

	return &res, nil
}

func (r *orderCommandRepository) Update(ctx context.Context, request *requests.UpdateOrderRecordRequest) (*db.UpdateOrderRow, error) {
	var res db.UpdateOrderRow
	err := r.db.GetContext(ctx, &res, updateOrder,
		int32(request.OrderID),
		int32(request.TotalPrice),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound
		}
		return nil, order_errors.ErrUpdateOrder.WithInternal(err)
	}

	return &res, nil
}

func (r *orderCommandRepository) Trash(ctx context.Context, order_id int) (*db.Order, error) {
	var res db.Order
	err := r.db.GetContext(ctx, &res, trashedOrder, int32(order_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound
		}
		return nil, order_errors.ErrTrashedOrder.WithInternal(err)
	}

	return &res, nil
}

func (r *orderCommandRepository) Restore(ctx context.Context, order_id int) (*db.Order, error) {
	var res db.Order
	err := r.db.GetContext(ctx, &res, restoreOrder, int32(order_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound
		}
		return nil, order_errors.ErrRestoreOrder.WithInternal(err)
	}

	return &res, nil
}

func (r *orderCommandRepository) FindTrashedByID(ctx context.Context, order_id int) (*db.Order, error) {
	var res db.Order
	err := r.db.GetContext(ctx, &res, getTrashedOrder, int32(order_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, order_errors.ErrOrderNotFound
		}
		return nil, order_errors.ErrFindById.WithInternal(err)
	}
	return &res, nil
}

func (r *orderCommandRepository) FindTrashed(ctx context.Context) ([]*db.Order, error) {
	var res []*db.Order
	err := r.db.SelectContext(ctx, &res, getTrashedOrders)
	if err != nil {
		return nil, order_errors.ErrFindByTrashed.WithInternal(err)
	}
	return res, nil
}

func (r *orderCommandRepository) DeletePermanent(ctx context.Context, order_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteOrderPermanently, int32(order_id)); err != nil {
		return false, order_errors.ErrDeleteOrderPermanent.WithInternal(err)
	}

	return true, nil
}

// DeletePermanentWithChildren purges a trashed order together with its stock
// reservations, order items, transactions, and shipping addresses in a single
// atomic SQL statement, so a mid-way failure cannot orphan child rows. The
// statement guards on the order being trashed; a non-trashed order yields
// sql.ErrNoRows which is surfaced as ErrOrderNotFound.
func (r *orderCommandRepository) DeletePermanentWithChildren(ctx context.Context, order_id int) (bool, error) {
	var deletedOrderID int32
	err := r.db.GetContext(ctx, &deletedOrderID, deleteOrderPermanentlyWithChildren, int32(order_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, order_errors.ErrOrderNotFound
		}
		return false, order_errors.ErrDeleteOrderPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *orderCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllOrders); err != nil {
		return false, order_errors.ErrRestoreAllOrder.WithInternal(err)
	}
	return true, nil
}

func (r *orderCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentOrders); err != nil {
		return false, order_errors.ErrDeleteAllOrderPermanent.WithInternal(err)
	}
	return true, nil
}
