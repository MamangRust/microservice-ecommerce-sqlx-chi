package merchantbusinessapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type merchantBusinessCommandResponseMapper struct{}

func NewMerchantBusinessCommandResponseMapper() MerchantBusinessCommandResponseMapper {
	return &merchantBusinessCommandResponseMapper{}
}

func (m *merchantBusinessCommandResponseMapper) ToResponseMerchantBusiness(merchant *pb_merchant_business.MerchantBusinessResponse) *response.MerchantBusinessResponse {
	if merchant == nil {
		return nil
	}
	return &response.MerchantBusinessResponse{
		ID:                int(merchant.Id),
		MerchantID:        int(merchant.MerchantId),
		BusinessType:      merchant.BusinessType,
		TaxID:             merchant.TaxId,
		EstablishedYear:   int(merchant.EstablishedYear),
		NumberOfEmployees: int(merchant.NumberOfEmployees),
		WebsiteUrl:        merchant.WebsiteUrl,
		MerchantName:      &merchant.MerchantName,
		CreatedAt:         merchant.CreatedAt,
		UpdatedAt:         merchant.UpdatedAt,
	}
}

func (m *merchantBusinessCommandResponseMapper) ToResponsesMerchantBusiness(merchants []*pb_merchant_business.MerchantBusinessResponse) []*response.MerchantBusinessResponse {
	var mappedMerchants []*response.MerchantBusinessResponse
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchantBusiness(merchant))
	}
	return mappedMerchants
}

func (m *merchantBusinessCommandResponseMapper) ToApiResponseMerchantBusiness(pbResponse *pb_merchant_business.ApiResponseMerchantBusiness) *response.ApiResponseMerchantBusiness {
	return &response.ApiResponseMerchantBusiness{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantBusiness(pbResponse.Data),
	}
}

func (m *merchantBusinessCommandResponseMapper) ToResponseMerchantBusinessDeleteAt(merchant *pb_merchant_business.MerchantBusinessResponseDeleteAt) *response.MerchantBusinessResponseDeleteAt {
	if merchant == nil {
		return nil
	}
	var deletedAt string
	if merchant.DeletedAt != nil {
		deletedAt = merchant.DeletedAt.Value
	}

	return &response.MerchantBusinessResponseDeleteAt{
		ID:                int(merchant.Id),
		MerchantID:        int(merchant.MerchantId),
		BusinessType:      merchant.BusinessType,
		TaxID:             merchant.TaxId,
		EstablishedYear:   int(merchant.EstablishedYear),
		NumberOfEmployees: int(merchant.NumberOfEmployees),
		WebsiteUrl:        merchant.WebsiteUrl,
		MerchantName:      merchant.MerchantName,
		CreatedAt:         merchant.CreatedAt,
		UpdatedAt:         merchant.UpdatedAt,
		DeletedAt:         &deletedAt,
	}
}

func (m *merchantBusinessCommandResponseMapper) ToResponsesMerchantBusinessDeleteAt(merchants []*pb_merchant_business.MerchantBusinessResponseDeleteAt) []*response.MerchantBusinessResponseDeleteAt {
	var mappedMerchants []*response.MerchantBusinessResponseDeleteAt
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchantBusinessDeleteAt(merchant))
	}
	return mappedMerchants
}

func (m *merchantBusinessCommandResponseMapper) ToApiResponseMerchantBusinessDeleteAt(pbResponse *pb_merchant_business.ApiResponseMerchantBusinessDeleteAt) *response.ApiResponseMerchantBusinessDeleteAt {
	return &response.ApiResponseMerchantBusinessDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantBusinessDeleteAt(pbResponse.Data),
	}
}
