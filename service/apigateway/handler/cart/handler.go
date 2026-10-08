package carthandler

import (
	cart_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/cart"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/cart"
	pbcart "github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsCart struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
}

func RegisterCartHandler(deps *DepsCart) {
	mapper := apimapper.NewCartResponseMapper()
	cache := cart_cache.NewCartMencache(deps.CacheStore)

	NewCartQueryHandleApi(&cartQueryHandleDeps{
		client: pbcart.NewCartQueryServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache,
	})

	NewCartCommandHandleApi(&cartCommandHandleDeps{
		client: pbcart.NewCartCommandServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.CommandMapper(),
		cache:  cache,
	})
}
