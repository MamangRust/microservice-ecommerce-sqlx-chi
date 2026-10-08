package producthandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	product_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/product"
	pbproduct "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/go-chi/chi/v5"
)

type productQueryHandlerApi struct {
	client pbproduct.ProductQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.ProductQueryResponseMapper
	cache  product_cache.ProductQueryCache
	errors apierror.ApiHandler
}

type productQueryHandleDeps struct {
	client     pbproduct.ProductQueryServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.ProductQueryResponseMapper
	cache      product_cache.ProductQueryCache
	apiHandler apierror.ApiHandler
}

func NewProductQueryHandleApi(params *productQueryHandleDeps) *productQueryHandlerApi {
	handler := &productQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
		errors: params.apiHandler,
	}

	params.router.Route("/api/product-query", func(routerProduct chi.Router) {
		routerProduct.Get("/", httpx.Handler(handler.FindAll))
		routerProduct.Get("/{id}", httpx.Handler(handler.FindById))
		routerProduct.Get("/merchant/{merchant_id}", httpx.Handler(handler.FindByMerchant))
		routerProduct.Get("/category/{category_name}", httpx.Handler(handler.FindByCategory))
		routerProduct.Get("/active", httpx.Handler(handler.FindByActive))
		routerProduct.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all products
// @Tags Product Query
// @Description Retrieve a list of all products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationProduct "List of products"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query [get]
func (h *productQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllProduct{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedProducts(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbproduct.FindAllProductRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationProduct(res)
	h.cache.SetCachedProducts(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find product by ID
// @Tags Product Query
// @Description Retrieve a product by ID
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.ApiResponseProduct "Product data"
// @Failure 400 {object} errors.ErrorResponse "Invalid product ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query/{id} [get]
func (h *productQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Product ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedProduct(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbproduct.FindByIdProductRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseProduct(res)
	h.cache.SetCachedProduct(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find products by merchant ID
// @Tags Product Query
// @Description Retrieve a list of products belonging to a specific merchant
// @Accept json
// @Produce json
// @Param merchant_id path int true "Merchant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationProduct "List of products by merchant"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query/merchant/{merchant_id} [get]
func (h *productQueryHandlerApi) FindByMerchant(w http.ResponseWriter, r *http.Request) error {
	merchantID, err := strconv.Atoi(chi.URLParam(r, "merchant_id"))
	if err != nil || merchantID <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Merchant ID")
	}

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
	req := &requests.FindAllProductByMerchant{MerchantID: merchantID, Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedProductsByMerchant(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByMerchant(ctx, &pbproduct.FindAllProductMerchantRequest{
		MerchantId: int32(merchantID), Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationProduct(res)
	h.cache.SetCachedProductsByMerchant(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find products by category name
// @Tags Product Query
// @Description Retrieve a list of products belonging to a specific category
// @Accept json
// @Produce json
// @Param category_name path string true "Category Name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationProduct "List of products by category"
// @Failure 400 {object} errors.ErrorResponse "Invalid category name"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query/category/{category_name} [get]
func (h *productQueryHandlerApi) FindByCategory(w http.ResponseWriter, r *http.Request) error {
	categoryName := chi.URLParam(r, "category_name")
	if categoryName == "" {
		return httpx.NewHTTPError(http.StatusBadRequest, "Category Name is required")
	}

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
	req := &requests.FindAllProductByCategory{CategoryName: categoryName, Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedProductsByCategory(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByCategory(ctx, &pbproduct.FindAllProductCategoryRequest{
		CategoryName: categoryName, Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationProduct(res)
	h.cache.SetCachedProductsByCategory(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active products
// @Tags Product Query
// @Description Retrieve a list of active products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationProductDeleteAt "List of active products"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query/active [get]
func (h *productQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllProduct{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedProductActive(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbproduct.FindAllProductRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationProductDeleteAt(res)
	h.cache.SetCachedProductActive(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed products
// @Tags Product Query
// @Description Retrieve a list of trashed product records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationProductDeleteAt "List of trashed product data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve product data"
// @Router /api/product-query/trashed [get]
func (h *productQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllProduct{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedProductTrashed(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbproduct.FindAllProductRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationProductDeleteAt(res)
	h.cache.SetCachedProductTrashed(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
