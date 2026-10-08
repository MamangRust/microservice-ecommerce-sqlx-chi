package merchanthandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchant_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/go-chi/chi/v5"
)

type merchantQueryHandlerApi struct {
	client     pbmerchant.MerchantQueryServiceClient
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantQueryResponseMapper
	cache      merchant_cache.MerchantQueryCache
	apiHandler apierror.ApiHandler
}

type merchantQueryHandleDeps struct {
	client     pbmerchant.MerchantQueryServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantQueryResponseMapper
	cache      merchant_cache.MerchantQueryCache
	apiHandler apierror.ApiHandler
}

func NewMerchantQueryHandleApi(params *merchantQueryHandleDeps) *merchantQueryHandlerApi {
	merchantQueryHandler := &merchantQueryHandlerApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/merchant-query", func(routerMerchant chi.Router) {

		routerMerchant.Get("/", httpx.Handler(merchantQueryHandler.FindAll))
		routerMerchant.Get("/{id}", httpx.Handler(merchantQueryHandler.FindById))
		routerMerchant.Get("/active", httpx.Handler(merchantQueryHandler.FindByActive))
		routerMerchant.Get("/trashed", httpx.Handler(merchantQueryHandler.FindByTrashed))

	})
	return merchantQueryHandler
}

// @Security Bearer
// @Summary Find all merchants
// @Tags Merchant Query
// @Description Retrieve a list of all merchants
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchant "List of merchants"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant data"
// @Router /api/merchant-query [get]
func (h *merchantQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchants(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchant(res)
	h.cache.SetCachedMerchants(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find merchant by ID
// @Tags Merchant Query
// @Description Retrieve a merchant by ID
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} response.ApiResponseMerchant "Merchant data"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant data"
// @Router /api/merchant-query/{id} [get]
func (h *merchantQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedMerchant(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseMerchant(res)
	h.cache.SetCachedMerchant(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active merchants
// @Tags Merchant Query
// @Description Retrieve a list of active merchants
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDeleteAt "List of active merchants"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant data"
// @Router /api/merchant-query/active [get]
func (h *merchantQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchantActive(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	// Note: MerchantQueryResponseMapper might need ToApiResponsePaginationMerchantDeleteAt
	// consolidated into it, similar to Category.
	apiResponse := h.mapper.ToApiResponsePaginationMerchantDeleteAt(res)
	h.cache.SetCachedMerchantActive(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed merchants
// @Tags Merchant Query
// @Description Retrieve a list of trashed merchant records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDeleteAt "List of trashed merchant data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant data"
// @Router /api/merchant-query/trashed [get]
func (h *merchantQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchantTrashed(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantDeleteAt(res)
	h.cache.SetCachedMerchantTrashed(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
