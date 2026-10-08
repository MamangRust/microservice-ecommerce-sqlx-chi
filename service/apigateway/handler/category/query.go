package categoryhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	category_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/category"
	pbcategory "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/go-chi/chi/v5"
)

type categoryQueryHandlerApi struct {
	client pbcategory.CategoryQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.CategoryQueryResponseMapper
	cache  category_cache.CategoryMencache
	errors apierror.ApiHandler
}

type categoryQueryHandleDeps struct {
	client     pbcategory.CategoryQueryServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.CategoryQueryResponseMapper
	cache      category_cache.CategoryMencache
	apiHandler apierror.ApiHandler
}

func NewCategoryQueryHandleApi(params *categoryQueryHandleDeps) *categoryQueryHandlerApi {
	categoryQueryHandler := &categoryQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
		errors: params.apiHandler,
	}

	params.router.Route("/api/category-query", func(routerCategory chi.Router) {

		routerCategory.Get("/", httpx.Handler(categoryQueryHandler.FindAll))
		routerCategory.Get("/{id}", httpx.Handler(categoryQueryHandler.FindById))
		routerCategory.Get("/active", httpx.Handler(categoryQueryHandler.FindByActive))
		routerCategory.Get("/trashed", httpx.Handler(categoryQueryHandler.FindByTrashed))

	})
	return categoryQueryHandler
}

// @Security Bearer
// @Summary Find all categories
// @Tags Category Query
// @Description Retrieve a list of all categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationCategory "List of categories"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve category data"
// @Router /api/category-query [get]
func (h *categoryQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedCategoriesCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbcategory.FindAllCategoryRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationCategory(res)
	h.cache.SetCachedCategoriesCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find category by ID
// @Tags Category Query
// @Description Retrieve a category by ID
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.ApiResponseCategory "Category data"
// @Failure 400 {object} errors.ErrorResponse "Invalid category ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve category data"
// @Router /api/category-query/{id} [get]
func (h *categoryQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedCategoryCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseCategory(res)
	h.cache.SetCachedCategoryCache(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active categories
// @Tags Category Query
// @Description Retrieve a list of active categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt "List of active categories"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve category data"
// @Router /api/category-query/active [get]
func (h *categoryQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedCategoryActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbcategory.FindAllCategoryRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationCategoryDeleteAt(res)
	h.cache.SetCachedCategoryActiveCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed categories
// @Tags Category Query
// @Description Retrieve a list of trashed category records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt "List of trashed category data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve category data"
// @Router /api/category-query/trashed [get]
func (h *categoryQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllCategory{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedCategoryTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbcategory.FindAllCategoryRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationCategoryDeleteAt(res)
	h.cache.SetCachedCategoryTrashedCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
