package orderhandler

import (
	"net/http"
	"strconv"

	order_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/order"
	pborder "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/go-chi/chi/v5"
)

type orderQueryHandlerApi struct {
	client pborder.OrderQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.OrderQueryResponseMapper
	cache  order_cache.OrderQueryCache
}

type orderQueryHandleDeps struct {
	client pborder.OrderQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.OrderQueryResponseMapper
	cache  order_cache.OrderQueryCache
}

func NewOrderQueryHandleApi(params *orderQueryHandleDeps) *orderQueryHandlerApi {
	handler := &orderQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/order-query", func(routerOrder chi.Router) {
		routerOrder.Get("/", httpx.Handler(handler.FindAll))
		routerOrder.Get("/{id}", httpx.Handler(handler.FindById))
		routerOrder.Get("/active", httpx.Handler(handler.FindByActive))
		routerOrder.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all orders
// @Tags Order Query
// @Description Retrieve a list of all orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrder "List of orders"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order data"
// @Router /api/order-query [get]
func (h *orderQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrder{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetOrderAllCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pborder.FindAllOrderRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrder(res)
	h.cache.SetOrderAllCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find order by ID
// @Tags Order Query
// @Description Retrieve an order by ID
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ApiResponseOrder "Order data"
// @Failure 400 {object} errors.ErrorResponse "Invalid order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order data"
// @Router /api/order-query/{id} [get]
func (h *orderQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedOrderCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pborder.FindByIdOrderRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseOrder(res)
	h.cache.SetCachedOrderCache(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active orders
// @Tags Order Query
// @Description Retrieve a list of active orders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrderDeleteAt "List of active orders"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order data"
// @Router /api/order-query/active [get]
func (h *orderQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrder{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetOrderActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pborder.FindAllOrderRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrderDeleteAt(res)
	h.cache.SetOrderActiveCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed orders
// @Tags Order Query
// @Description Retrieve a list of trashed order records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrderDeleteAt "List of trashed order data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order data"
// @Router /api/order-query/trashed [get]
func (h *orderQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrder{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetOrderTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pborder.FindAllOrderRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrderDeleteAt(res)
	h.cache.SetOrderTrashedCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
