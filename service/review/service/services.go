package service

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-review/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	ReviewQuery   ReviewQueryService
	ReviewCommand ReviewCommandService
}

type Deps struct {
	Observability observability.TraceLoggerObservability
	Cache         *cache.Mencache
	Repositories  *repository.Repositories
	Logger        logger.LoggerInterface
}

func NewService(deps *Deps) *Service {
	return &Service{
		ReviewQuery: NewReviewQueryService(&ReviewQueryServiceDeps{
			Observability:    deps.Observability,
			Cache:            deps.Cache.ReviewQuery,
			ReviewRepository: deps.Repositories.ReviewQuery,
			Logger:           deps.Logger,
		}),
		ReviewCommand: NewReviewCommandService(&ReviewCommandServiceDeps{
			Observability:         deps.Observability,
			Cache:                 deps.Cache.ReviewCommand,
			ReviewRepository:      deps.Repositories.ReviewCommand,
			ReviewQueryRepository: deps.Repositories.ReviewQuery,
			UserRepository:        deps.Repositories.UserQuery,
			ProductRepository:     deps.Repositories.ProductQuery,
			Logger:                deps.Logger,
		}),
	}
}
