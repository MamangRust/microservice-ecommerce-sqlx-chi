package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_social_link"
)

type MerchantDetailQueryHandler interface {
	pb_merchant_detail.MerchantDetailQueryServiceServer
}

type MerchantDetailCommandHandler interface {
	pb_merchant_detail.MerchantDetailCommandServiceServer
}

type MerchantSocialLinkCommandHandler interface {
	pb_merchant_social_link.MerchantSocialCommandServiceServer
}
