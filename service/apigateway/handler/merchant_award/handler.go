package merchantawardhandler

import (
	merchantaward_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_awards"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_award"
	pbmerchantaward "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_award"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsMerchantAward struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
}

func RegisterMerchantAwardHandler(deps *DepsMerchantAward) {
	mapper := apimapper.NewMerchantAwardResponseMapper()
	merchantMapper := merchantapimapper.NewMerchantResponseMapper()
	cache := merchantaward_cache.NewMerchantAward(deps.CacheStore)

	NewMerchantAwardQueryHandleApi(&merchantAwardQueryHandleDeps{
		client: pbmerchantaward.NewMerchantAwardQueryServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache,
	})

	NewMerchantAwardCommandHandleApi(&merchantAwardCommandHandleDeps{
		client:         pbmerchantaward.NewMerchantAwardCommandServiceClient(deps.Client),
		router:         deps.Router,
		logger:         deps.Logger,
		mapper:         mapper.CommandMapper(),
		merchantMapper: merchantMapper.CommandMapper(),
		cache:          cache,
	})
}
