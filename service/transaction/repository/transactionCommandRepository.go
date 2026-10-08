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

const createTransaction = `INSERT INTO transactions (
    order_id, merchant_id, payment_method, amount, payment_status
) VALUES ($1, $2, $3, $4, $5)
RETURNING transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at`

const updateTransaction = `UPDATE transactions
SET merchant_id = $2,
    payment_method = $3,
    amount = $4,
    payment_status = $5,
    order_id = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE transaction_id = $1
  AND deleted_at IS NULL
RETURNING transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at`

const trashTransaction = `UPDATE transactions
SET
    deleted_at = current_timestamp
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
    RETURNING transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at,
    deleted_at`

const restoreTransaction = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    transaction_id = $1
    AND deleted_at IS NOT NULL
  RETURNING transaction_id,
    order_id,
    merchant_id,
    payment_method,
    amount,
    payment_status,
    created_at,
    updated_at,
    deleted_at`

const deleteTransactionPermanently = `DELETE FROM transactions WHERE transaction_id = $1 AND deleted_at IS NOT NULL`

const deleteTransactionByOrderPermanent = `DELETE FROM transactions
WHERE
    order_id = $1`

const restoreAllTransactions = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentTransactions = `DELETE FROM transactions
WHERE
    deleted_at IS NOT NULL`

type transactionCommandRepository struct {
	db *sqlx.DB
}

func NewTransactionCommandRepository(db *sqlx.DB) *transactionCommandRepository {
	return &transactionCommandRepository{
		db: db,
	}
}

func (r *transactionCommandRepository) Create(ctx context.Context, request *requests.CreateTransactionRequest) (*db.CreateTransactionRow, error) {
	var row db.CreateTransactionRow

	if err := r.db.GetContext(ctx, &row, createTransaction,
		int32(request.OrderID),
		int32(request.MerchantID),
		request.PaymentMethod,
		int32(request.Amount),
		*request.PaymentStatus,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound
		}
		return nil, transaction_errors.ErrCreateTransaction.WithInternal(err)
	}

	return &row, nil
}

// CreateInTx runs the transaction insert inside the given database transaction so
// the caller can commit the business row and its outbox event atomically.
func (r *transactionCommandRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, request *requests.CreateTransactionRequest) (*db.CreateTransactionRow, error) {
	var row db.CreateTransactionRow

	if err := tx.GetContext(ctx, &row, createTransaction,
		int32(request.OrderID),
		int32(request.MerchantID),
		request.PaymentMethod,
		int32(request.Amount),
		*request.PaymentStatus,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound
		}
		return nil, transaction_errors.ErrCreateTransaction.WithInternal(err)
	}

	return &row, nil
}

func (r *transactionCommandRepository) Update(ctx context.Context, request *requests.UpdateTransactionRequest) (*db.UpdateTransactionRow, error) {
	var row db.UpdateTransactionRow

	if err := r.db.GetContext(ctx, &row, updateTransaction,
		int32(*request.TransactionID),
		int32(request.MerchantID),
		request.PaymentMethod,
		int32(request.Amount),
		*request.PaymentStatus,
		int32(request.OrderID),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound
		}
		return nil, transaction_errors.ErrUpdateTransaction.WithInternal(err)
	}

	return &row, nil
}

func (r *transactionCommandRepository) Trash(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var row db.Transaction

	if err := r.db.GetContext(ctx, &row, trashTransaction, int32(transaction_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound
		}
		return nil, transaction_errors.ErrTrashTransaction.WithInternal(err)
	}

	return &row, nil
}

func (r *transactionCommandRepository) Restore(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var row db.Transaction

	if err := r.db.GetContext(ctx, &row, restoreTransaction, int32(transaction_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, transaction_errors.ErrTransactionNotFound
		}
		return nil, transaction_errors.ErrRestoreTransaction.WithInternal(err)
	}

	return &row, nil
}

func (r *transactionCommandRepository) DeletePermanent(ctx context.Context, transaction_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteTransactionPermanently, int32(transaction_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, transaction_errors.ErrTransactionNotFound
		}
		return false, transaction_errors.ErrDeleteTransactionPermanently.WithInternal(err)
	}

	return true, nil
}

func (r *transactionCommandRepository) DeleteByOrderIDPermanent(ctx context.Context, order_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteTransactionByOrderPermanent, int32(order_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, transaction_errors.ErrTransactionNotFound
		}
		return false, transaction_errors.ErrDeleteTransactionPermanently.WithInternal(err)
	}

	return true, nil
}

func (r *transactionCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllTransactions); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, transaction_errors.ErrTransactionNotFound
		}
		return false, transaction_errors.ErrRestoreAllTransactions.WithInternal(err)
	}

	return true, nil
}

func (r *transactionCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentTransactions); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, transaction_errors.ErrTransactionNotFound
		}
		return false, transaction_errors.ErrDeleteAllTransactionPermanent.WithInternal(err)
	}

	return true, nil
}
