package shippingaddresshandler

import (
	shippingaddress_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/shipping_address"
	pbshippingaddress "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsShippingAddress struct {
	Client *grpc.ClientConn
	Router chi.Router
	Logger logger.LoggerInterface
	Cache  *cache.CacheStore
}

func RegisterShippingAddressHandler(deps *DepsShippingAddress) {
	mapper := apimapper.NewShippingAddressResponseMapper()
	cache := shippingaddress_cache.NewShippingAddressMencache(deps.Cache)

	NewShippingAddressQueryHandleApi(&shippingAddressQueryHandleDeps{
		client: pbshippingaddress.NewShippingQueryServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache.QueryCache(),
	})

	NewShippingAddressCommandHandleApi(&shippingAddressCommandHandleDeps{
		client: pbshippingaddress.NewShippingCommandServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.CommandMapper(),
		cache:  cache.CommandCache(),
	})
}
