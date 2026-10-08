package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-product/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Handler struct {
	ProductQuery   ProductQueryHandler
	ProductCommand ProductCommandHandler
}

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		ProductQuery:   NewProductQueryHandler(deps.Service.ProductQuery, deps.Logger),
		ProductCommand: NewProductCommandHandler(deps.Service.ProductCommand, deps.Logger),
	}
}
