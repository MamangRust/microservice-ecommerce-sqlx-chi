package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	ShippingQuery   pb_shipping_address.ShippingQueryServiceServer
	ShippingCommand pb_shipping_address.ShippingCommandServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		ShippingQuery:   NewShippingQueryHandler(deps.Service.ShippingAddressQuery, deps.Logger),
		ShippingCommand: NewShippingCommandHandler(deps.Service.ShippingAddressCommand, deps.Logger),
	}
}
