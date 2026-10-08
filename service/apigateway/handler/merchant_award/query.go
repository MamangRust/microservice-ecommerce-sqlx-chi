package merchantawardhandler

import (
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"net/http"
	"strconv"

	merchantawards_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_awards"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_award"
	pbmerchantaward "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_award"
	"github.com/go-chi/chi/v5"
)

type merchantAwardQueryHandlerApi struct {
	client pbmerchantaward.MerchantAwardQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.MerchantAwardQueryResponseMapper
	cache  merchantawards_cache.MerchantAwardQueryCache
}

type merchantAwardQueryHandleDeps struct {
	client pbmerchantaward.MerchantAwardQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.MerchantAwardQueryResponseMapper
	cache  merchantawards_cache.MerchantAwardQueryCache
}

func NewMerchantAwardQueryHandleApi(params *merchantAwardQueryHandleDeps) *merchantAwardQueryHandlerApi {
	handler := &merchantAwardQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/merchant-award-query", func(routerAward chi.Router) {
		routerAward.Get("/", httpx.Handler(handler.FindAll))
		routerAward.Get("/{id}", httpx.Handler(handler.FindById))
		routerAward.Get("/active", httpx.Handler(handler.FindByActive))
		routerAward.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all merchant awards
// @Tags Merchant Award Query
// @Description Retrieve a list of all merchant awards/certifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantAward "List of merchant awards"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant award data"
// @Router /api/merchant-award-query [get]
func (h *merchantAwardQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchantAwardAll(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantAward(res)
	h.cache.SetCachedMerchantAwardAll(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find merchant award by ID
// @Tags Merchant Award Query
// @Description Retrieve a merchant award by ID
// @Accept json
// @Produce json
// @Param id path int true "Award ID"
// @Success 200 {object} response.ApiResponseMerchantAward "Merchant award data"
// @Failure 400 {object} errors.ErrorResponse "Invalid award ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant award data"
// @Router /api/merchant-award-query/{id} [get]
func (h *merchantAwardQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedMerchantAward(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbmerchantaward.FindByIdMerchantAwardRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseMerchantAward(res)
	h.cache.SetCachedMerchantAward(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active merchant awards
// @Tags Merchant Award Query
// @Description Retrieve a list of active merchant awards
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantAwardDeleteAt "List of active merchant awards"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant award data"
// @Router /api/merchant-award-query/active [get]
func (h *merchantAwardQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchantAwardActive(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantAwardDeleteAt(res)
	h.cache.SetCachedMerchantAwardActive(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed merchant awards
// @Tags Merchant Award Query
// @Description Retrieve a list of trashed merchant award records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantAwardDeleteAt "List of trashed merchant award data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant award data"
// @Router /api/merchant-award-query/trashed [get]
func (h *merchantAwardQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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

	if cachedData, found := h.cache.GetCachedMerchantAwardTrashed(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantAwardDeleteAt(res)
	h.cache.SetCachedMerchantAwardTrashed(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
