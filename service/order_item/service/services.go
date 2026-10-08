package service

import (
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-order-item/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-order-item/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	OrderItemQuery   OrderItemQueryService
	OrderItemCommand OrderItemCommandService
}

type Deps struct {
	Repository    *repository.Repositories
	Cache         *mencache.Mencache
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		OrderItemQuery: NewOrderItemQueryService(&OrderItemQueryServiceDeps{
			Observability:       deps.Observability,
			Cache:               deps.Cache.OrderItemQueryCache,
			OrderItemRepository: deps.Repository.OrderItemQuery,
			Logger:              deps.Logger,
		}),
		OrderItemCommand: NewOrderItemCommandService(&OrderItemCommandServiceDeps{
			Observability:       deps.Observability,
			Cache:               deps.Cache.OrderItemCommandCache,
			OrderItemRepository: deps.Repository.OrderItemCommand,
			Logger:              deps.Logger,
		}),
	}
}
