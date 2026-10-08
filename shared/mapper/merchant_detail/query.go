package merchantdetailapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type merchantDetailQueryResponseMapper struct {
	MerchantDetailCommandResponseMapper
}

func NewMerchantDetailQueryResponseMapper() MerchantDetailQueryResponseMapper {
	return &merchantDetailQueryResponseMapper{
		MerchantDetailCommandResponseMapper: NewMerchantDetailCommandResponseMapper(),
	}
}

func (m *merchantDetailQueryResponseMapper) ToResponseMerchantDetail(merchant *pb_merchant_detail.MerchantDetailResponse) *response.MerchantDetailResponse {
	return m.MerchantDetailCommandResponseMapper.ToResponseMerchantDetail(merchant)
}

func (m *merchantDetailQueryResponseMapper) ToResponseMerchantDetailRelation(merchant *pb_merchant_detail.MerchantDetailResponse) *response.MerchantDetailResponse {
	return m.MerchantDetailCommandResponseMapper.ToResponseMerchantDetailRelation(merchant)
}

func (m *merchantDetailQueryResponseMapper) ToResponsesMerchantDetail(merchants []*pb_merchant_detail.MerchantDetailResponse) []*response.MerchantDetailResponse {
	return m.MerchantDetailCommandResponseMapper.ToResponsesMerchantDetail(merchants)
}

func (m *merchantDetailQueryResponseMapper) ToApiResponseMerchantDetail(pbResponse *pb_merchant_detail.ApiResponseMerchantDetail) *response.ApiResponseMerchantDetail {
	return m.MerchantDetailCommandResponseMapper.ToApiResponseMerchantDetail(pbResponse)
}

func (m *merchantDetailQueryResponseMapper) ToApiResponseMerchantDetailRelation(pbResponse *pb_merchant_detail.ApiResponseMerchantDetail) *response.ApiResponseMerchantDetailRelation {
	return &response.ApiResponseMerchantDetailRelation{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantDetailRelation(pbResponse.Data),
	}
}

func (m *merchantDetailQueryResponseMapper) ToApiResponsesMerchantDetail(pbResponse *pb_merchant_detail.ApiResponsesMerchantDetail) *response.ApiResponsesMerchantDetail {
	return &response.ApiResponsesMerchantDetail{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponsesMerchantDetail(pbResponse.Data),
	}
}

func (m *merchantDetailQueryResponseMapper) ToApiResponsePaginationMerchantDetail(pbResponse *pb_merchant_detail.ApiResponsePaginationMerchantDetail) *response.ApiResponsePaginationMerchantDetail {
	return &response.ApiResponsePaginationMerchantDetail{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       m.ToResponsesMerchantDetail(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}

func (m *merchantDetailQueryResponseMapper) ToApiResponsePaginationMerchantDetailDeleteAt(pbResponse *pb_merchant_detail.ApiResponsePaginationMerchantDetailDeleteAt) *response.ApiResponsePaginationMerchantDetailDeleteAt {
	return m.MerchantDetailCommandResponseMapper.ToApiResponsePaginationMerchantDetailDeleteAt(pbResponse)
}
