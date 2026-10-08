package service

import (
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-product/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-product/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	ProductQuery   ProductQueryService
	ProductCommand ProductCommandService
}

type Deps struct {
	Cache         *mencache.Mencache
	Repository    *repository.Repositories
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		ProductQuery: NewProductQueryService(&ProductQueryServiceDeps{
			Observability:     deps.Observability,
			Cache:             deps.Cache.ProductQuery,
			ProductRepository: deps.Repository.ProductQuery,
			Logger:            deps.Logger,
		}),
		ProductCommand: NewProductCommandService(&ProductCommandServiceDeps{
			Observability:      deps.Observability,
			Cache:              deps.Cache.ProductCommand,
			CategoryRepository: deps.Repository.CategoryQuery,
			MerchantRepository: deps.Repository.MerchantQuery,
			ProductQueryRepo:   deps.Repository.ProductQuery,
			ProductRepository:  deps.Repository.ProductCommand,
			Logger:             deps.Logger,
		}),
	}
}
