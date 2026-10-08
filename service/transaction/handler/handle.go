package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

// F5: legacy OLTP transaction stats handlers were removed; stats are served by
// service/stats_reader from ClickHouse.
type Handler struct {
	TransactionQuery   TransactionQueryHandler
	TransactionCommand TransactionCommandHandler
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		TransactionQuery:   NewTransactionQueryHandler(deps.Service.TransactionQuery, deps.Logger),
		TransactionCommand: NewTransactionCommandHandler(deps.Service.TransactionCommand, deps.Logger),
	}
}
