package merchantpolicyhandler

import (
	merchantpolicy_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_policies"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbmerchantpolicy "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsMerchantPolicy struct {
	Client        *grpc.ClientConn
	Router        chi.Router
	Logger        logger.LoggerInterface
	CacheStore    *cache.CacheStore
	Observability observability.TraceLoggerObservability
}

func RegisterMerchantPolicyHandler(deps *DepsMerchantPolicy) {
	mapper := apimapper.NewMerchantPolicyResponseMapper()
	merchantMapper := merchantapimapper.NewMerchantResponseMapper()
	cache := merchantpolicy_cache.NewMerchantPoliciesMencache(deps.CacheStore)

	NewMerchantPolicyQueryHandleApi(&merchantPolicyQueryHandleDeps{
		client:        pbmerchantpolicy.NewMerchantPolicyQueryServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.QueryMapper(),
		cache:         cache,
		observability: deps.Observability,
	})

	NewMerchantPolicyCommandHandleApi(&merchantPolicyCommandHandleDeps{
		client:         pbmerchantpolicy.NewMerchantPolicyCommandServiceClient(deps.Client),
		router:         deps.Router,
		logger:         deps.Logger,
		mapper:         mapper.CommandMapper(),
		merchantMapper: merchantMapper.CommandMapper(),
		cache:          cache,
		observability:  deps.Observability,
	})
}
