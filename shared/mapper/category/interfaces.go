package categoryapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type CategoryBaseResponseMapper interface {
	ToResponseCategory(category *pb_category.CategoryResponse) *response.CategoryResponse
	ToResponsesCategory(categories []*pb_category.CategoryResponse) []*response.CategoryResponse
}

type CategoryQueryResponseMapper interface {
	CategoryBaseResponseMapper
	ToApiResponseCategory(pbResponse *pb_category.ApiResponseCategory) *response.ApiResponseCategory
	ToApiResponsesCategory(pbResponse *pb_category.ApiResponsesCategory) *response.ApiResponsesCategory
	ToApiResponsePaginationCategory(pbResponse *pb_category.ApiResponsePaginationCategory) *response.ApiResponsePaginationCategory
	ToApiResponsePaginationCategoryDeleteAt(pbResponse *pb_category.ApiResponsePaginationCategoryDeleteAt) *response.ApiResponsePaginationCategoryDeleteAt
}

type CategoryCommandResponseMapper interface {
	CategoryBaseResponseMapper
	ToResponseCategoryDelete(category *pb_category.CategoryResponseDeleteAt) *response.CategoryResponseDeleteAt
	ToResponsesCategoryDeleteAt(categories []*pb_category.CategoryResponseDeleteAt) []*response.CategoryResponseDeleteAt
	ToApiResponseCategoryDeleteAt(pbResponse *pb_category.ApiResponseCategoryDeleteAt) *response.ApiResponseCategoryDeleteAt
	ToApiResponseCategory(pbResponse *pb_category.ApiResponseCategory) *response.ApiResponseCategory
	ToApiResponseCategoryDelete(pbResponse *pb_category.ApiResponseCategoryDelete) *response.ApiResponseCategoryDelete
	ToApiResponseCategoryAll(pbResponse *pb_category.ApiResponseCategoryAll) *response.ApiResponseCategoryAll
	ToApiResponsePaginationCategoryDeleteAt(pbResponse *pb_category.ApiResponsePaginationCategoryDeleteAt) *response.ApiResponsePaginationCategoryDeleteAt
}
