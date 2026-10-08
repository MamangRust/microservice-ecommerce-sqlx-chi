package service

import (
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	MerchantPoliciesQuery   MerchantPoliciesQueryService
	MerchantPoliciesCommand MerchantPoliciesCommandService
}

type Deps struct {
	Cache         *mencache.Mencache
	Repository    *repository.Repositories
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		MerchantPoliciesQuery: NewMerchantPoliciesQueryService(&MerchantPoliciesQueryServiceDeps{
			Observability:            deps.Observability,
			Cache:                    deps.Cache.MerchantPoliciesQueryCache,
			MerchantPolicyRepository: deps.Repository.MerchantPoliciesQuery,
			Logger:                   deps.Logger,
		}),
		MerchantPoliciesCommand: NewMerchantPoliciesCommandService(&MerchantPoliciesCommandServiceDeps{
			Observability:            deps.Observability,
			Cache:                    deps.Cache.MerchantPoliciesCommandCache,
			MerchantPolicyRepository: deps.Repository.MerchantPoliciesCommand,
			Logger:                   deps.Logger,
		}),
	}
}
