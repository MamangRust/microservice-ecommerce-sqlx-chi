package service

import (
	mencache "github.com/MamangRust/microservice-ecommerce-grpc-transaction/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/jmoiron/sqlx"
)

// F5: legacy OLTP transaction stats services were removed; stats are served by
// service/stats_reader from ClickHouse.
type Service struct {
	TransactionQuery   TransactionQueryService
	TransactionCommand TransactionCommandService
	Outbox             OutboxService
}

type Deps struct {
	Kafka         *kafka.Kafka
	Pool          *sqlx.DB
	Cache         *mencache.Mencache
	Repositories  *repository.Repositories
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

func NewService(deps *Deps) *Service {
	return &Service{
		TransactionQuery: NewTransactionQueryService(&TransactionQueryServiceDeps{
			Observability: deps.Observability,
			Cache:         deps.Cache.TransactionQueryCache,
			Repository:    deps.Repositories.TransactionQuery,
			Logger:        deps.Logger,
		}),
		TransactionCommand: NewTransactionCommandService(&TransactionCommandServiceDeps{
			Observability:      deps.Observability,
			Kafka:              deps.Kafka,
			Pool:               deps.Pool,
			Cache:              deps.Cache.TransactionCommandCache,
			TransactionQuery:   deps.Repositories.TransactionQuery,
			TransactionCommand: deps.Repositories.TransactionCommand,
			UserQuery:          deps.Repositories.UserQuery,
			MerchantQuery:      deps.Repositories.MerchantQuery,
			OrderQuery:         deps.Repositories.OrderQuery,
			OrderItem:          deps.Repositories.OrderItem,
			ShippingAddress:    deps.Repositories.ShippingAddress,
			Outbox:             deps.Repositories.Outbox,
			Logger:             deps.Logger,
		}),
		Outbox: NewOutboxService(deps.Repositories.Outbox, deps.Kafka, deps.Logger),
	}
}
