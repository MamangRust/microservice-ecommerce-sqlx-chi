package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	MerchantQuery           pb_merchant.MerchantQueryServiceServer
	MerchantCommandHandler  pb_merchant.MerchantCommandServiceServer
	MerchantDocumentQuery   pb_merchant_document.MerchantDocumentQueryServiceServer
	MerchantDocumentCommand pb_merchant_document.MerchantDocumentCommandServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		MerchantQuery:           NewMerchantQueryHandler(deps.Service.MerchantQuery, deps.Logger),
		MerchantCommandHandler:  NewMerchantCommandHandler(deps.Service.MerchantCommand, deps.Logger),
		MerchantDocumentQuery:   NewMerchantDocumentQueryHandler(deps.Service.MerchantDocumentQuery, deps.Logger),
		MerchantDocumentCommand: NewMerchantDocumentCommandHandler(deps.Service.MerchantDocumentCommand, deps.Logger),
	}
}
