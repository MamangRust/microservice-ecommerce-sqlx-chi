package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateTransactionRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetTransactionByIDRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetTransactionByOrderIDRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetTransactionsRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetTransactionsActiveRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetTransactionsTrashedRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type GetTransactionByMerchantRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
}

type UpdateTransactionRow struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
