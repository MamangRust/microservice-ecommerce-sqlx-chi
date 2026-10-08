package category_cache

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type CategoryQueryCache interface {
	GetCachedCategoriesCache(ctx context.Context, req *requests.FindAllCategory) (*response.ApiResponsePaginationCategory, bool)
	SetCachedCategoriesCache(ctx context.Context, req *requests.FindAllCategory, data *response.ApiResponsePaginationCategory)
	GetCachedCategoryActiveCache(ctx context.Context, req *requests.FindAllCategory) (*response.ApiResponsePaginationCategoryDeleteAt, bool)
	SetCachedCategoryActiveCache(ctx context.Context, req *requests.FindAllCategory, data *response.ApiResponsePaginationCategoryDeleteAt)
	GetCachedCategoryTrashedCache(ctx context.Context, req *requests.FindAllCategory) (*response.ApiResponsePaginationCategoryDeleteAt, bool)
	SetCachedCategoryTrashedCache(ctx context.Context, req *requests.FindAllCategory, data *response.ApiResponsePaginationCategoryDeleteAt)
	GetCachedCategoryCache(ctx context.Context, id int) (*response.ApiResponseCategory, bool)
	SetCachedCategoryCache(ctx context.Context, data *response.ApiResponseCategory)
}

type CategoryCommandCache interface {
	DeleteCachedCategoryCache(ctx context.Context, id int)
}
