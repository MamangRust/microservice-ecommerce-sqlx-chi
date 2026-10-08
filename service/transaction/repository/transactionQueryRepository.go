package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-transaction/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/transaction_errors"
	"github.com/jmoiron/sqlx"
)

const getTransactions = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM transactions
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR payment_method ILIKE '%' || $1 || '%' OR payment_status ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getTransactionsActive = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM transactions
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR payment_method ILIKE '%' || $1 || '%' OR payment_status ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getTransactionsTrashed = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM transactions
WHERE deleted_at IS NOT NULL
AND ($1::TEXT IS NULL OR payment_method ILIKE '%' || $1 || '%' OR payment_status ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getTransactionByMerchant = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM transactions
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR payment_method ILIKE '%' || $1 || '%' OR payment_status ILIKE '%' || $1 || '%')
  AND ($2::INT IS NULL OR merchant_id = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

const getTransactionByOrderID = `SELECT
    transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at
FROM transactions
WHERE order_id = $1
  AND deleted_at IS NULL`

const getTransactionByID = `SELECT transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at
FROM transactions
WHERE transaction_id = $1
  AND deleted_at IS NULL`

type transactionQueryRepository struct {
	db *sqlx.DB
}

func NewTransactionQueryRepository(db *sqlx.DB) *transactionQueryRepository {
	return &transactionQueryRepository{
		db: db,
	}
}

func (r *transactionQueryRepository) FindAll(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetTransactionsRow

	if err := r.db.SelectContext(ctx, &rows, getTransactions,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, transaction_errors.ErrFindAllTransactions.WithInternal(err)
	}

	return rows, nil
}

func (r *transactionQueryRepository) FindActive(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetTransactionsActiveRow

	if err := r.db.SelectContext(ctx, &rows, getTransactionsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, transaction_errors.ErrFindByActive.WithInternal(err)
	}

	return rows, nil
}

func (r *transactionQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetTransactionsTrashedRow

	if err := r.db.SelectContext(ctx, &rows, getTransactionsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, transaction_errors.ErrFindByTrashed.WithInternal(err)
	}

	return rows, nil
}

func (r *transactionQueryRepository) FindByMerchant(
	ctx context.Context,
	req *requests.FindAllTransactionByMerchant,
) ([]*db.GetTransactionByMerchantRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetTransactionByMerchantRow

	if err := r.db.SelectContext(ctx, &rows, getTransactionByMerchant,
		req.Search,
		int32(req.MerchantID),
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, transaction_errors.ErrFindByMerchant.WithInternal(err)
	}

	return rows, nil
}

func (r *transactionQueryRepository) FindByID(ctx context.Context, transaction_id int) (*db.GetTransactionByIDRow, error) {
	var row db.GetTransactionByIDRow

	if err := r.db.GetContext(ctx, &row, getTransactionByID, int32(transaction_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound.WithInternal(err)
		}
		return nil, transaction_errors.ErrFindById.WithInternal(err)
	}

	return &row, nil
}

func (r *transactionQueryRepository) FindByOrderID(ctx context.Context, order_id int) (*db.GetTransactionByOrderIDRow, error) {
	var row db.GetTransactionByOrderIDRow

	if err := r.db.GetContext(ctx, &row, getTransactionByOrderID, int32(order_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound.WithInternal(err)
		}
		return nil, transaction_errors.ErrFindByOrderId.WithInternal(err)
	}

	return &row, nil
}
