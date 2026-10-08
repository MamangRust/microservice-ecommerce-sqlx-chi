package productapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type ProductBaseResponseMapper interface {
	ToResponseProduct(product *pb_product.ProductResponse) *response.ProductResponse
	ToResponsesProduct(products []*pb_product.ProductResponse) []*response.ProductResponse
	ToResponseProductDeleteAt(product *pb_product.ProductResponseDeleteAt) *response.ProductResponseDeleteAt
	ToResponsesProductDeleteAt(products []*pb_product.ProductResponseDeleteAt) []*response.ProductResponseDeleteAt
	ToApiResponseProduct(pbResponse *pb_product.ApiResponseProduct) *response.ApiResponseProduct
	ToApiResponsePaginationProductDeleteAt(pbResponse *pb_product.ApiResponsePaginationProductDeleteAt) *response.ApiResponsePaginationProductDeleteAt
}

type ProductQueryResponseMapper interface {
	ProductBaseResponseMapper
	ToApiResponsesProduct(pbResponse *pb_product.ApiResponsesProduct) *response.ApiResponsesProduct
	ToApiResponsePaginationProduct(pbResponse *pb_product.ApiResponsePaginationProduct) *response.ApiResponsePaginationProduct
}

type ProductCommandResponseMapper interface {
	ProductBaseResponseMapper
	ToApiResponsesProductDeleteAt(pbResponse *pb_product.ApiResponseProductDeleteAt) *response.ApiResponseProductDeleteAt
	ToApiResponseProductDelete(pbResponse *pb_product.ApiResponseProductDelete) *response.ApiResponseProductDelete
	ToApiResponseProductAll(pbResponse *pb_product.ApiResponseProductAll) *response.ApiResponseProductAll
}
