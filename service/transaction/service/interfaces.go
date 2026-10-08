package service

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-transaction/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

// F5: legacy OLTP transaction stats services were removed; stats are served by
// service/stats_reader from ClickHouse.

type TransactionQueryService interface {
	FindAll(
		ctx context.Context,
		req *requests.FindAllTransaction,
	) ([]*db.GetTransactionsRow, *int, error)

	FindActive(
		ctx context.Context,
		req *requests.FindAllTransaction,
	) ([]*db.GetTransactionsActiveRow, *int, error)

	FindTrashed(
		ctx context.Context,
		req *requests.FindAllTransaction,
	) ([]*db.GetTransactionsTrashedRow, *int, error)

	FindByMerchant(
		ctx context.Context,
		req *requests.FindAllTransactionByMerchant,
	) ([]*db.GetTransactionByMerchantRow, *int, error)

	FindByID(
		ctx context.Context,
		transaction_id int,
	) (*db.GetTransactionByIDRow, error)

	FindByOrderID(
		ctx context.Context,
		order_id int,
	) (*db.GetTransactionByOrderIDRow, error)
}

type TransactionCommandService interface {
	Create(
		ctx context.Context,
		request *requests.CreateTransactionRequest,
	) (*db.CreateTransactionRow, error)

	Update(
		ctx context.Context,
		request *requests.UpdateTransactionRequest,
	) (*db.UpdateTransactionRow, error)

	Trash(
		ctx context.Context,
		transaction_id int,
	) (*db.Transaction, error)

	Restore(
		ctx context.Context,
		transaction_id int,
	) (*db.Transaction, error)

	DeletePermanent(
		ctx context.Context,
		transaction_id int,
	) (bool, error)

	DeleteByOrderIDPermanent(
		ctx context.Context,
		order_id int,
	) (bool, error)

	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}
