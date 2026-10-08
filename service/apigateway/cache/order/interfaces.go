package order_cache

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type OrderQueryCache interface {
	GetOrderAllCache(ctx context.Context, req *requests.FindAllOrder) (*response.ApiResponsePaginationOrder, bool)
	SetOrderAllCache(ctx context.Context, req *requests.FindAllOrder, data *response.ApiResponsePaginationOrder)

	GetOrderActiveCache(ctx context.Context, req *requests.FindAllOrder) (*response.ApiResponsePaginationOrderDeleteAt, bool)
	SetOrderActiveCache(ctx context.Context, req *requests.FindAllOrder, data *response.ApiResponsePaginationOrderDeleteAt)

	GetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrder) (*response.ApiResponsePaginationOrderDeleteAt, bool)
	SetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrder, data *response.ApiResponsePaginationOrderDeleteAt)

	GetCachedOrderCache(ctx context.Context, order_id int) (*response.ApiResponseOrder, bool)
	SetCachedOrderCache(ctx context.Context, data *response.ApiResponseOrder)
}

type OrderCommandCache interface {
	DeleteOrderCache(ctx context.Context, orderID int)
}
