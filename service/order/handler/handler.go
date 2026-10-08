package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-order/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

// F5: legacy OLTP order stats handler was removed; stats are served by
// service/stats_reader from ClickHouse.
type Handler struct {
	OrderQuery   OrderQueryHandler
	OrderCommand OrderCommandHandler
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		OrderQuery:   NewOrderQueryHandler(deps.Service.OrderQuery, deps.Logger),
		OrderCommand: NewOrderCommandHandler(deps.Service.OrderCommand, deps.Logger),
	}
}
