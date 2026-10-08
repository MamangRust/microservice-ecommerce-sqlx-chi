package cache

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-transaction/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)



type TransactionQueryCache interface {
	GetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsRow, *int, bool)
	SetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction, data []*db.GetTransactionsRow, total *int)

	GetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*db.GetTransactionByMerchantRow, *int, bool)
	SetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant, data []*db.GetTransactionByMerchantRow, total *int)

	GetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsActiveRow, *int, bool)
	SetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction, data []*db.GetTransactionsActiveRow, total *int)

	GetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsTrashedRow, *int, bool)
	SetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction, data []*db.GetTransactionsTrashedRow, total *int)

	GetCachedTransactionCache(ctx context.Context, id int) (*db.GetTransactionByIDRow, bool)
	SetCachedTransactionCache(ctx context.Context, data *db.GetTransactionByIDRow)

	GetCachedTransactionByOrderId(ctx context.Context, orderID int) (*db.GetTransactionByOrderIDRow, bool)
	SetCachedTransactionByOrderId(ctx context.Context, orderID int, data *db.GetTransactionByOrderIDRow)
}

type TransactionCommandCache interface {
	DeleteTransactionCache(ctx context.Context, transactionID int)
	InvalidateTransactionCache(ctx context.Context)
}
