package categoryapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type categoryQueryResponseMapper struct {
	CategoryCommandResponseMapper
}

func NewCategoryQueryResponseMapper() CategoryQueryResponseMapper {
	return &categoryQueryResponseMapper{
		CategoryCommandResponseMapper: NewCategoryCommandResponseMapper(),
	}
}

func (c *categoryQueryResponseMapper) ToResponseCategory(category *pb_category.CategoryResponse) *response.CategoryResponse {
	return &response.CategoryResponse{
		ID:            int(category.Id),
		Name:          category.Name,
		Description:   category.Description,
		SlugCategory:  category.SlugCategory,
		ImageCategory: category.ImageCategory,
		CreatedAt:     category.CreatedAt,
		UpdatedAt:     category.UpdatedAt,
	}
}

func (c *categoryQueryResponseMapper) ToResponsesCategory(categories []*pb_category.CategoryResponse) []*response.CategoryResponse {
	var mappedCategories []*response.CategoryResponse
	for _, category := range categories {
		mappedCategories = append(mappedCategories, c.ToResponseCategory(category))
	}
	return mappedCategories
}

func (c *categoryQueryResponseMapper) ToApiResponseCategory(pbResponse *pb_category.ApiResponseCategory) *response.ApiResponseCategory {
	return &response.ApiResponseCategory{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCategory(pbResponse.Data),
	}
}

func (c *categoryQueryResponseMapper) ToApiResponsesCategory(pbResponse *pb_category.ApiResponsesCategory) *response.ApiResponsesCategory {
	return &response.ApiResponsesCategory{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponsesCategory(pbResponse.Data),
	}
}

func (c *categoryQueryResponseMapper) ToApiResponsePaginationCategory(pbResponse *pb_category.ApiResponsePaginationCategory) *response.ApiResponsePaginationCategory {
	return &response.ApiResponsePaginationCategory{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCategory(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}
