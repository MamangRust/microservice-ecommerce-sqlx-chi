package service

import (
	merchantCache "github.com/MamangRust/microservice-ecommerce-grpc-merchant/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/outbox"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	MerchantQuery           MerchantQueryService
	MerchantCommand         MerchantCommandService
	MerchantDocumentCommand MerchantDocumentCommandService
	MerchantDocumentQuery   MerchantDocumentQueryService
}

type Deps struct {
	Kafka         *kafka.Kafka
	Repositories  *repository.Repositories
	Mencache      *merchantCache.Mencache
	Pool          *sqlx.DB
	Outbox        *outbox.OutboxService
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		MerchantQuery: NewMerchantQueryService(&MerchantQueryServiceDeps{
			Cache:              deps.Mencache.MerchantQueryCache,
			MerchantRepository: deps.Repositories.MerchantQuery,
			Logger:             deps.Logger,
			Observability:      deps.Observability,
		}),
		MerchantCommand: NewMerchantCommandService(&MerchantCommandServiceDeps{
			Kafka:              deps.Kafka,
			Cache:              deps.Mencache.MerchantCommandCache,
			MerchantRepository: deps.Repositories.MerchantCommand,
			MerchantQuery:      deps.Repositories.MerchantQuery,
			UserRepository:     deps.Repositories.UserQuery,
			Pool:               deps.Pool,
			Outbox:             deps.Outbox,
			Logger:             deps.Logger,
			Observability:      deps.Observability,
		}),
		MerchantDocumentCommand: NewMerchantDocumentCommandService(&MerchantDocumentCommandServiceDeps{
			Kafka:         deps.Kafka,
			Cache:         deps.Mencache.MerchantDocumentCommandCache,
			Repository:    deps.Repositories.MerchantDocumentCommand,
			MerchantQuery: deps.Repositories.MerchantQuery,
			UserQuery:     deps.Repositories.UserQuery,
			Pool:          deps.Pool,
			Outbox:        deps.Outbox,
			Logger:        deps.Logger,
			Observability: deps.Observability,
		}),
		MerchantDocumentQuery: NewMerchantDocumentQueryService(&MerchantDocumentQueryServiceDeps{
			Cache:         deps.Mencache.MerchantDocumentQueryCache,
			Repository:    deps.Repositories.MerchantDocumentQuery,
			Logger:        deps.Logger,
			Observability: deps.Observability,
		}),
	}
}
