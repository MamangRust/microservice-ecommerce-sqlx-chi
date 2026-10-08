package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-slider/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	SliderQuery   pb_slider.SliderQueryServiceServer
	SliderCommand pb_slider.SliderCommandServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		SliderQuery:   NewSliderQueryHandler(deps.Service.SliderQuery, deps.Logger),
		SliderCommand: NewSliderCommandHandler(deps.Service.SliderCommand, deps.Logger),
	}
}
