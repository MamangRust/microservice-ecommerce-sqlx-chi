package service

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-user/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	UserQuery   UserQueryService
	UserCommand UserCommandService
}

type Deps struct {
	Cache         cache.UserMencache
	Repositories  *repository.Repositories
	Hash          hash.HashPassword
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		UserQuery: NewUserQueryService(&UserQueryServiceDeps{
			Observability:  deps.Observability,
			Cache:          deps.Cache,
			UserRepository: deps.Repositories.UserQuery,
			Logger:         deps.Logger,
		}),
		UserCommand: NewUserCommandService(&UserCommandServiceDeps{
			Observability:         deps.Observability,
			Cache:                 deps.Cache,
			UserCommandRepository: deps.Repositories.UserCommand,
			UserQueryRepository:   deps.Repositories.UserQuery,
			RoleRepository:        deps.Repositories.Role,
			UserRoleRepository:    deps.Repositories.UserRole,
			Logger:                deps.Logger,
			Hash:                  deps.Hash,
		}),
	}
}
