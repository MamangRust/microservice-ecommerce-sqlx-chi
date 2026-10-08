package cache

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)



type OrderQueryCache interface {
	GetOrderAllCache(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersRow, *int, bool)
	SetOrderAllCache(ctx context.Context, req *requests.FindAllOrder, data []*db.GetOrdersRow, total *int)

	GetOrderActiveCache(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersActiveRow, *int, bool)
	SetOrderActiveCache(ctx context.Context, req *requests.FindAllOrder, data []*db.GetOrdersActiveRow, total *int)

	GetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrder) ([]*db.GetOrdersTrashedRow, *int, bool)
	SetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrder, data []*db.GetOrdersTrashedRow, total *int)

	GetCachedOrderCache(ctx context.Context, orderID int) (*db.GetOrderByIDRow, bool)
	SetCachedOrderCache(ctx context.Context, data *db.GetOrderByIDRow)

	GetOrderByMerchantCache(ctx context.Context, req *requests.FindAllOrderByMerchant) ([]*db.GetOrdersByMerchantRow, *int, bool)
	SetOrderByMerchantCache(ctx context.Context, req *requests.FindAllOrderByMerchant, data []*db.GetOrdersByMerchantRow, total *int)
}

type OrderCommandCache interface {
	DeleteOrderCache(ctx context.Context, orderID int)
	InvalidateOrderCache(ctx context.Context)
}
