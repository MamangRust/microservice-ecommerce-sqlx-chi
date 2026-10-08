package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order-item/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/order_item_errors"
	"github.com/jmoiron/sqlx"
)

const createOrderItem = `INSERT INTO
    order_items (
        order_id,
        product_id,
        quantity,
        price
    )
VALUES ($1, $2, $3, $4)
RETURNING
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at`

const updateOrderItem = `UPDATE order_items
SET
    quantity = $2,
    price = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE
    order_item_id = $1
    AND deleted_at IS NULL
RETURNING
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at`

const trashOrderItem = `UPDATE order_items
SET
    deleted_at = current_timestamp
WHERE
    order_item_id = $1
    AND deleted_at IS NULL
RETURNING
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at,
    deleted_at`

const restoreOrderItem = `UPDATE order_items
SET
    deleted_at = NULL
WHERE
    order_item_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at,
    deleted_at`

const deleteOrderItemPermanently = `DELETE FROM order_items
WHERE
    order_item_id = $1
    AND deleted_at IS NOT NULL`

const deleteOrderItemsByOrderPermanent = `DELETE FROM order_items
WHERE
    order_id = $1`

const restoreAllOrdersItem = `UPDATE order_items
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentOrdersItem = `DELETE FROM order_items
WHERE
    deleted_at IS NOT NULL`

const calculateTotalPrice = `SELECT COALESCE(SUM(quantity * price), 0)::int AS total_price
FROM order_items
WHERE
    order_id = $1
    AND deleted_at IS NULL`

type orderItemCommandRepository struct {
	db *sqlx.DB
}

func NewOrderItemCommandRepository(db *sqlx.DB) *orderItemCommandRepository {
	return &orderItemCommandRepository{
		db: db,
	}
}

func (r *orderItemCommandRepository) Create(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*db.CreateOrderItemRow, error) {
	var item db.CreateOrderItemRow
	err := r.db.GetContext(ctx, &item, createOrderItem,
		int32(req.OrderID),
		int32(req.ProductID),
		int32(req.Quantity),
		int32(req.Price),
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderItemCommandRepository) Update(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*db.UpdateOrderItemRow, error) {
	var item db.UpdateOrderItemRow
	err := r.db.GetContext(ctx, &item, updateOrderItem,
		int32(req.OrderItemID),
		int32(req.Quantity),
		int32(req.Price),
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderItemCommandRepository) Trash(ctx context.Context, orderItemID int) (*db.OrderItem, error) {
	var item db.OrderItem
	err := r.db.GetContext(ctx, &item, trashOrderItem, int32(orderItemID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orderitem_errors.ErrOrderItemNotFound
		}
		return nil, orderitem_errors.ErrTrashedOrderItem
	}
	return &item, nil
}

func (r *orderItemCommandRepository) Restore(ctx context.Context, orderItemID int) (*db.OrderItem, error) {
	var item db.OrderItem
	err := r.db.GetContext(ctx, &item, restoreOrderItem, int32(orderItemID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orderitem_errors.ErrOrderItemNotFound
		}
		return nil, orderitem_errors.ErrRestoreOrderItem
	}
	return &item, nil
}

func (r *orderItemCommandRepository) DeletePermanent(ctx context.Context, orderItemID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteOrderItemPermanently, int32(orderItemID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, orderitem_errors.ErrOrderItemNotFound
		}
		return false, orderitem_errors.ErrDeleteOrderItemPermanent
	}
	return true, nil
}

func (r *orderItemCommandRepository) DeleteOrderItemByOrderPermanent(ctx context.Context, orderID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteOrderItemsByOrderPermanent, int32(orderID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, orderitem_errors.ErrOrderItemNotFound
		}
		return false, orderitem_errors.ErrDeleteOrderItemPermanent
	}
	return true, nil
}

func (r *orderItemCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllOrdersItem); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, orderitem_errors.ErrOrderItemNotFound
		}
		return false, orderitem_errors.ErrRestoreAllOrderItem
	}
	return true, nil
}

func (r *orderItemCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentOrdersItem); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, orderitem_errors.ErrOrderItemNotFound
		}
		return false, orderitem_errors.ErrDeleteAllOrderPermanent
	}
	return true, nil
}

func (r *orderItemCommandRepository) CalculateTotalPrice(ctx context.Context, orderID int) (int, error) {
	var total int32
	err := r.db.GetContext(ctx, &total, calculateTotalPrice, int32(orderID))
	if err != nil {
		return 0, orderitem_errors.ErrCalculateTotalPrice
	}
	return int(total), nil
}
