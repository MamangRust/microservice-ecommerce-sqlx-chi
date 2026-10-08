package merchantsociallinkapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_social_link"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type MerchantSocialLinkBaseResponseMapper interface {
	MapMerchantSocialLink(doc *pb_merchant_detail.MerchantSocialMediaLinkResponse) *response.MerchantSocialLinkResponse
	ToApiResponseMerchantSocialLink(doc *pb_merchant_social_link.ApiResponseMerchantSocial) *response.ApiResponseMerchantSocialLink
}

type MerchantSocialLinkQueryResponseMapper interface {
	MerchantSocialLinkBaseResponseMapper
}

type MerchantSocialLinkCommandResponseMapper interface {
	MerchantSocialLinkBaseResponseMapper
}
