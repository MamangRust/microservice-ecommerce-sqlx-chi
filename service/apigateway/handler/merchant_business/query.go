package merchantbusinesshandler

import (
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"net/http"
	"strconv"

	merchantbusiness_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_business"
	pbmerchantbusiness "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/go-chi/chi/v5"
)

type merchantBusinessQueryHandlerApi struct {
	client pbmerchantbusiness.MerchantBusinessQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.MerchantBusinessQueryResponseMapper
	cache  merchantbusiness_cache.MerchantBusinessQueryCache
}

type merchantBusinessQueryHandleDeps struct {
	client pbmerchantbusiness.MerchantBusinessQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.MerchantBusinessQueryResponseMapper
	cache  merchantbusiness_cache.MerchantBusinessQueryCache
}

func NewMerchantBusinessQueryHandleApi(params *merchantBusinessQueryHandleDeps) *merchantBusinessQueryHandlerApi {
	handler := &merchantBusinessQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/merchant-business-query", func(routerBusiness chi.Router) {
		routerBusiness.Get("/", httpx.Handler(handler.FindAll))
		routerBusiness.Get("/{id}", httpx.Handler(handler.FindById))
		routerBusiness.Get("/active", httpx.Handler(handler.FindByActive))
		routerBusiness.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all merchant business details
// @Tags Merchant Business Query
// @Description Retrieve a list of all merchant business details
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantBusiness "List of merchant business details"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant business data"
// @Router /api/merchant-business-query [get]
func (h *merchantBusinessQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantBusinessAll(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantBusiness(res)
	h.cache.SetCachedMerchantBusinessAll(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find merchant business by ID
// @Tags Merchant Business Query
// @Description Retrieve a merchant business detail by ID
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 200 {object} response.ApiResponseMerchantBusiness "Merchant business data"
// @Failure 400 {object} errors.ErrorResponse "Invalid business ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant business data"
// @Router /api/merchant-business-query/{id} [get]
func (h *merchantBusinessQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedMerchantBusiness(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbmerchantbusiness.FindByIdMerchantBusinessRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseMerchantBusiness(res)
	h.cache.SetCachedMerchantBusiness(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active merchant business details
// @Tags Merchant Business Query
// @Description Retrieve a list of active merchant business details
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantBusinessDeleteAt "List of active merchant business details"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant business data"
// @Router /api/merchant-business-query/active [get]
func (h *merchantBusinessQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantBusinessActive(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantBusinessDeleteAt(res)
	h.cache.SetCachedMerchantBusinessActive(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed merchant business details
// @Tags Merchant Business Query
// @Description Retrieve a list of trashed merchant business records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantBusinessDeleteAt "List of trashed merchant business data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant business data"
// @Router /api/merchant-business-query/trashed [get]
func (h *merchantBusinessQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantBusinessTrashed(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantBusinessDeleteAt(res)
	h.cache.SetCachedMerchantBusinessTrashed(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
