package merchantsociallinkapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_social_link"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type merchantSocialLinkCommandResponseMapper struct{}

func NewMerchantSocialLinkCommandResponseMapper() MerchantSocialLinkCommandResponseMapper {
	return &merchantSocialLinkCommandResponseMapper{}
}

func (m *merchantSocialLinkCommandResponseMapper) MapMerchantSocialLink(doc *pb_merchant_detail.MerchantSocialMediaLinkResponse) *response.MerchantSocialLinkResponse {
	if doc == nil {
		return nil
	}
	return &response.MerchantSocialLinkResponse{
		ID:               int(doc.Id),
		MerchantDetailID: int(doc.MerchantDetailId),
		Platform:         doc.Platform,
		URL:              doc.Url,
	}
}

func (m *merchantSocialLinkCommandResponseMapper) ToApiResponseMerchantSocialLink(doc *pb_merchant_social_link.ApiResponseMerchantSocial) *response.ApiResponseMerchantSocialLink {
	return &response.ApiResponseMerchantSocialLink{
		Status:  doc.Status,
		Message: doc.Message,
		Data:    m.MapMerchantSocialLink(doc.Data),
	}
}
