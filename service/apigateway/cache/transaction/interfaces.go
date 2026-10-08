package transaction_cache

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type TransactionQueryCache interface {
	GetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction) (*response.ApiResponsePaginationTransaction, bool)
	SetCachedTransactionsCache(ctx context.Context, req *requests.FindAllTransaction, data *response.ApiResponsePaginationTransaction)

	GetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) (*response.ApiResponsePaginationTransaction, bool)
	SetCachedTransactionByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant, data *response.ApiResponsePaginationTransaction)

	GetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction) (*response.ApiResponsePaginationTransaction, bool)
	SetCachedTransactionActiveCache(ctx context.Context, req *requests.FindAllTransaction, data *response.ApiResponsePaginationTransaction)

	GetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction) (*response.ApiResponsePaginationTransactionDeleteAt, bool)
	SetCachedTransactionTrashedCache(ctx context.Context, req *requests.FindAllTransaction, data *response.ApiResponsePaginationTransactionDeleteAt)

	GetCachedTransactionCache(ctx context.Context, id int) (*response.ApiResponseTransaction, bool)
	SetCachedTransactionCache(ctx context.Context, data *response.ApiResponseTransaction)

	GetCachedTransactionByOrderId(ctx context.Context, orderID int) (*response.ApiResponseTransaction, bool)
	SetCachedTransactionByOrderId(ctx context.Context, orderID int, data *response.ApiResponseTransaction)
}

type TransactionCommandCache interface {
	DeleteTransactionCache(ctx context.Context, transactionID int)
	InvalidateTransactionCache(ctx context.Context)
}
