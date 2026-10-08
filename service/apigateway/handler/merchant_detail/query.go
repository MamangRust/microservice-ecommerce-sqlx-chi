package merchantdetailhandler

import (
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchant_detail_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_detail"
	pbmerchantdetail "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/go-chi/chi/v5"
)

type merchantDetailQueryHandlerApi struct {
	client     pbmerchantdetail.MerchantDetailQueryServiceClient
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantDetailQueryResponseMapper
	cache      merchant_detail_cache.MerchantDetailQueryCache
	apiHandler apierror.ApiHandler
}

type merchantDetailQueryHandleDeps struct {
	client     pbmerchantdetail.MerchantDetailQueryServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantDetailQueryResponseMapper
	cache      merchant_detail_cache.MerchantDetailQueryCache
	apiHandler apierror.ApiHandler
}

func NewMerchantDetailQueryHandleApi(params *merchantDetailQueryHandleDeps) *merchantDetailQueryHandlerApi {
	handler := &merchantDetailQueryHandlerApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/merchant-detail-query", func(router chi.Router) {

		router.Get("/", httpx.Handler(handler.FindAll))
		router.Get("/{id}", httpx.Handler(handler.FindById))
		router.Get("/active", httpx.Handler(handler.FindByActive))
		router.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all merchant details
// @Tags Merchant Detail Query
// @Description Retrieve a list of all merchant details
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDetail "List of merchant details"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant detail data"
// @Router /api/merchant-detail-query [get]
func (h *merchantDetailQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	cacheReq := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantDetailAll(ctx, cacheReq); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{Page: int32(page), PageSize: int32(pageSize), Search: search})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantDetail(res)
	h.cache.SetCachedMerchantDetailAll(ctx, cacheReq, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find merchant detail by ID
// @Tags Merchant Detail Query
// @Description Retrieve a merchant detail by ID with all relations
// @Accept json
// @Produce json
// @Param id path int true "Detail ID"
// @Success 200 {object} response.ApiResponseMerchantDetailRelation "Merchant detail data"
// @Failure 400 {object} errors.ErrorResponse "Invalid detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant detail data"
// @Router /api/merchant-detail-query/{id} [get]
func (h *merchantDetailQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return sharedErrors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedMerchantDetailRelation(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbmerchantdetail.FindByIdMerchantDetailRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseMerchantDetailRelation(res)
	h.cache.SetCachedMerchantDetailRelation(ctx, id, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active merchant details
// @Tags Merchant Detail Query
// @Description Retrieve a list of active merchant details
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDetailDeleteAt "List of active merchant details"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant detail data"
// @Router /api/merchant-detail-query/active [get]
func (h *merchantDetailQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	cacheReq := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantDetailActive(ctx, cacheReq); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{Page: int32(page), PageSize: int32(pageSize), Search: search})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantDetailDeleteAt(res)
	h.cache.SetCachedMerchantDetailActive(ctx, cacheReq, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed merchant details
// @Tags Merchant Detail Query
// @Description Retrieve a list of trashed merchant detail records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDetailDeleteAt "List of trashed merchant detail data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant detail data"
// @Router /api/merchant-detail-query/trashed [get]
func (h *merchantDetailQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	cacheReq := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedMerchantDetailTrashed(ctx, cacheReq); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{Page: int32(page), PageSize: int32(pageSize), Search: search})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantDetailDeleteAt(res)
	h.cache.SetCachedMerchantDetailTrashed(ctx, cacheReq, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
