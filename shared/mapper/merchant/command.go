package merchantapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type merchantCommandResponseMapper struct{}

func NewMerchantCommandResponseMapper() MerchantCommandResponseMapper {
	return &merchantCommandResponseMapper{}
}

func (m *merchantCommandResponseMapper) ToResponseMerchant(merchant *pb_merchant.MerchantResponse) *response.MerchantResponse {
	return &response.MerchantResponse{
		ID:           int(merchant.Id),
		UserID:       int(merchant.UserId),
		Name:         merchant.Name,
		Description:  merchant.Description,
		Address:      merchant.Address,
		ContactEmail: merchant.ContactEmail,
		ContactPhone: merchant.ContactPhone,
		Status:       merchant.Status,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
	}
}

func (m *merchantCommandResponseMapper) ToResponsesMerchant(merchants []*pb_merchant.MerchantResponse) []*response.MerchantResponse {
	var mappedMerchants []*response.MerchantResponse
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchant(merchant))
	}
	return mappedMerchants
}

func (m *merchantCommandResponseMapper) ToApiResponseMerchant(pbResponse *pb_merchant.ApiResponseMerchant) *response.ApiResponseMerchant {
	return &response.ApiResponseMerchant{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchant(pbResponse.Data),
	}
}

func (m *merchantCommandResponseMapper) ToResponseMerchantDeleteAt(merchant *pb_merchant.MerchantResponseDeleteAt) *response.MerchantResponseDeleteAt {
	var deletedAt string
	if merchant.DeletedAt != nil {
		deletedAt = merchant.DeletedAt.Value
	}

	return &response.MerchantResponseDeleteAt{
		ID:           int(merchant.Id),
		UserID:       int(merchant.UserId),
		Name:         merchant.Name,
		Description:  merchant.Description,
		Address:      merchant.Address,
		ContactEmail: merchant.ContactEmail,
		ContactPhone: merchant.ContactPhone,
		Status:       merchant.Status,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
		DeletedAt:    &deletedAt,
	}
}

func (m *merchantCommandResponseMapper) ToResponsesMerchantDeleteAt(merchants []*pb_merchant.MerchantResponseDeleteAt) []*response.MerchantResponseDeleteAt {
	var mappedMerchants []*response.MerchantResponseDeleteAt
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchantDeleteAt(merchant))
	}
	return mappedMerchants
}

func (m *merchantCommandResponseMapper) ToApiResponseMerchantDeleteAt(pbResponse *pb_merchant.ApiResponseMerchantDeleteAt) *response.ApiResponseMerchantDeleteAt {
	return &response.ApiResponseMerchantDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantDeleteAt(pbResponse.Data),
	}
}

func (m *merchantCommandResponseMapper) ToApiResponseMerchantDelete(pbResponse *pb_merchant.ApiResponseMerchantDelete) *response.ApiResponseMerchantDelete {
	return &response.ApiResponseMerchantDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (m *merchantCommandResponseMapper) ToApiResponseMerchantAll(pbResponse *pb_merchant.ApiResponseMerchantAll) *response.ApiResponseMerchantAll {
	return &response.ApiResponseMerchantAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (m *merchantCommandResponseMapper) ToApiResponsePaginationMerchantDeleteAt(pbResponse *pb_merchant.ApiResponsePaginationMerchantDeleteAt) *response.ApiResponsePaginationMerchantDeleteAt {
	return &response.ApiResponsePaginationMerchantDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       m.ToResponsesMerchantDeleteAt(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}
