package merchantapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type MerchantBaseResponseMapper interface {
	ToResponseMerchant(merchant *pb_merchant.MerchantResponse) *response.MerchantResponse
	ToResponsesMerchant(merchants []*pb_merchant.MerchantResponse) []*response.MerchantResponse
}

type MerchantQueryResponseMapper interface {
	MerchantBaseResponseMapper
	ToApiResponseMerchant(pbResponse *pb_merchant.ApiResponseMerchant) *response.ApiResponseMerchant
	ToApiResponsesMerchant(pbResponse *pb_merchant.ApiResponsesMerchant) *response.ApiResponsesMerchant
	ToApiResponsePaginationMerchant(pbResponse *pb_merchant.ApiResponsePaginationMerchant) *response.ApiResponsePaginationMerchant
	ToApiResponsePaginationMerchantDeleteAt(pbResponse *pb_merchant.ApiResponsePaginationMerchantDeleteAt) *response.ApiResponsePaginationMerchantDeleteAt
}

type MerchantCommandResponseMapper interface {
	MerchantBaseResponseMapper
	ToApiResponseMerchant(pbResponse *pb_merchant.ApiResponseMerchant) *response.ApiResponseMerchant
	ToResponseMerchantDeleteAt(merchant *pb_merchant.MerchantResponseDeleteAt) *response.MerchantResponseDeleteAt
	ToResponsesMerchantDeleteAt(merchants []*pb_merchant.MerchantResponseDeleteAt) []*response.MerchantResponseDeleteAt
	ToApiResponseMerchantDeleteAt(pbResponse *pb_merchant.ApiResponseMerchantDeleteAt) *response.ApiResponseMerchantDeleteAt
	ToApiResponseMerchantDelete(pbResponse *pb_merchant.ApiResponseMerchantDelete) *response.ApiResponseMerchantDelete
	ToApiResponseMerchantAll(pbResponse *pb_merchant.ApiResponseMerchantAll) *response.ApiResponseMerchantAll
	ToApiResponsePaginationMerchantDeleteAt(pbResponse *pb_merchant.ApiResponsePaginationMerchantDeleteAt) *response.ApiResponsePaginationMerchantDeleteAt
}
