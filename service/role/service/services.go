package service

import (
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-role/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-role/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
)

type Service struct {
	RoleQuery   RoleQueryService
	RoleCommand RoleCommandService
}

type Deps struct {
	Cache         mencache.RoleMencache
	Repository    *repository.Repositories
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		RoleQuery: NewRoleQueryService(&RoleQueryServiceDeps{
			Observability:  deps.Observability,
			Cache:          deps.Cache,
			RoleRepository: deps.Repository.RoleQuery,
			Logger:         deps.Logger,
		}),
		RoleCommand: NewRoleCommandService(&RoleCommandServiceDeps{
			Observability:      deps.Observability,
			Cache:              deps.Cache,
			RoleRepository:     deps.Repository.RoleCommand,
			UserRoleRepository: deps.Repository.UserRole,
			Logger:             deps.Logger,
		}),
	}
}
