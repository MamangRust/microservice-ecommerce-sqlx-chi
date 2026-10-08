package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	MerchantDetailQuery      MerchantDetailQueryHandler
	MerchantDetailCommand    MerchantDetailCommandHandler
	MerchantSocialLinkCommand MerchantSocialLinkCommandHandler
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		MerchantDetailQuery:      NewMerchantDetailQueryHandler(deps.Service.MerchantDetailQuery, deps.Logger),
		MerchantDetailCommand:    NewMerchantDetailCommandHandler(deps.Service.MerchantDetailCommand, deps.Logger),
		MerchantSocialLinkCommand: NewMerchantSocialLinkCommandHandler(deps.Service.MerchantSocialLinkCommand, deps.Logger),
	}
}
