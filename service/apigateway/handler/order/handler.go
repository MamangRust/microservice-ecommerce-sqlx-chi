package orderhandler

import (
	order_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/order"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/order"
	pborder "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsOrder struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
}

func RegisterOrderHandler(deps *DepsOrder) {
	mapper := apimapper.NewOrderResponseMapper()
	cache := order_cache.OrderNewMencache(deps.CacheStore)

	queryClient := pborder.NewOrderQueryServiceClient(deps.Client)

	NewOrderQueryHandleApi(&orderQueryHandleDeps{
		client: queryClient,
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache,
	})

	NewOrderCommandHandleApi(&orderCommandHandleDeps{
		client: pborder.NewOrderCommandServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.CommandMapper(),
		cache:  cache,
	})
}
