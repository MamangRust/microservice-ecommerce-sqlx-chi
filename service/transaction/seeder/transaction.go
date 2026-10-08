package seeder

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const getTransactions = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM transactions
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR payment_method ILIKE '%' || $1 || '%' OR payment_status ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

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

type transactionSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewTransactionSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *transactionSeeder {
	return &transactionSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *transactionSeeder) Seed() error {
	// Idempotency: skip when transactions already exist.
	var existing []struct {
		TransactionID int32 `db:"transaction_id"`
		TotalCount    int64 `db:"total_count"`
	}

	if err := r.db.SelectContext(r.ctx, &existing, getTransactions, "", 1, 0); err == nil && len(existing) > 0 {
		r.logger.Debug("transactions already seeded, skipping")
		return nil
	}

	transactions := []struct {
		OrderID       int32
		MerchantID    int32
		PaymentMethod string
		Amount        int32
		PaymentStatus string
	}{
		{OrderID: 1, MerchantID: 1, PaymentMethod: "credit_card", Amount: 120000, PaymentStatus: "success"},
		{OrderID: 2, MerchantID: 2, PaymentMethod: "bank_transfer", Amount: 45000, PaymentStatus: "success"},
		{OrderID: 3, MerchantID: 3, PaymentMethod: "ewallet", Amount: 78900, PaymentStatus: "success"},
		{OrderID: 4, MerchantID: 4, PaymentMethod: "cash_on_delivery", Amount: 32000, PaymentStatus: "success"},
		{OrderID: 5, MerchantID: 5, PaymentMethod: "credit_card", Amount: 99999, PaymentStatus: "failed"},
		{OrderID: 6, MerchantID: 6, PaymentMethod: "bank_transfer", Amount: 150000, PaymentStatus: "failed"},
		{OrderID: 7, MerchantID: 7, PaymentMethod: "ewallet", Amount: 51000, PaymentStatus: "failed"},
		{OrderID: 8, MerchantID: 8, PaymentMethod: "credit_card", Amount: 67500, PaymentStatus: "failed"},
	}

	for _, tx := range transactions {
		if _, err := r.db.ExecContext(r.ctx, createTransaction,
			tx.OrderID,
			tx.MerchantID,
			tx.PaymentMethod,
			tx.Amount,
			tx.PaymentStatus,
		); err != nil {
			r.logger.Error("failed to seed transaction", zap.Error(err))
			return err
		}
	}

	r.logger.Info("transaction successfully seeded")

	return nil
}
