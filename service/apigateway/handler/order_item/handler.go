package orderitemhandler

import (
	orderitem_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/order_item"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/order_item"
	pborderitem "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsOrderItem struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
}

func RegisterOrderItemHandler(deps *DepsOrderItem) {
	mapper := apimapper.NewOrderItemResponseMapper()
	cache := orderitem_cache.NewOrderItemMencache(deps.CacheStore)

	queryClient := pborderitem.NewOrderItemQueryServiceClient(deps.Client)

	NewOrderItemQueryHandleApi(&orderItemQueryHandleDeps{
		client: queryClient,
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache,
	})

	NewOrderItemCommandHandleApi(&orderItemCommandHandleDeps{
		client: pborderitem.NewOrderItemCommandServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.CommandMapper(),
		cache:  cache,
	})
}
