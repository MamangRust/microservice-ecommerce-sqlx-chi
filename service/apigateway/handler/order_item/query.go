package orderitemhandler

import (
	"net/http"
	"strconv"

	orderitem_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/order_item"
	pborderitem "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/go-chi/chi/v5"
)

type orderItemQueryHandlerApi struct {
	client pborderitem.OrderItemQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.OrderItemQueryResponseMapper
	cache  orderitem_cache.OrderItemQueryCache
}

type orderItemQueryHandleDeps struct {
	client pborderitem.OrderItemQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.OrderItemQueryResponseMapper
	cache  orderitem_cache.OrderItemQueryCache
}

func NewOrderItemQueryHandleApi(params *orderItemQueryHandleDeps) *orderItemQueryHandlerApi {
	handler := &orderItemQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	// /api/order-item is also mounted by the command handler; chi does not
	// allow mounting the same path twice, so query routes register on the
	// parent router with full paths.
	routerOrderItem := params.router
	routerOrderItem.Get("/api/order-item", httpx.Handler(handler.FindAll))
	routerOrderItem.Get("/api/order-item/order/{order_id}", httpx.Handler(handler.FindOrderItemByOrder))
	routerOrderItem.Get("/api/order-item/active", httpx.Handler(handler.FindByActive))
	routerOrderItem.Get("/api/order-item/trashed", httpx.Handler(handler.FindByTrashed))
	return handler
}

// @Security Bearer
// @Summary Find all order items
// @Tags Order Item Query
// @Description Retrieve a list of all order items
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrderItem "List of order items"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order item data"
// @Router /api/order-item [get]
func (h *orderItemQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrderItems{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedOrderItemsAll(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindAll")
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrderItem(res)
	h.cache.SetCachedOrderItemsAll(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active order items
// @Tags Order Item Query
// @Description Retrieve a list of active order items
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrderItemDeleteAt "List of active order items"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order item data"
// @Router /api/order-item/active [get]
func (h *orderItemQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrderItems{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedOrderItemActive(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pborderitem.FindAllOrderItemRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindByActive")
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrderItemDeleteAt(res)
	h.cache.SetCachedOrderItemActive(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed order items
// @Tags Order Item Query
// @Description Retrieve a list of trashed order item records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationOrderItemDeleteAt "List of trashed order item data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order item data"
// @Router /api/order-item/trashed [get]
func (h *orderItemQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllOrderItems{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedOrderItemTrashed(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pborderitem.FindAllOrderItemRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindByTrashed")
	}

	apiResponse := h.mapper.ToApiResponsePaginationOrderItemDeleteAt(res)
	h.cache.SetCachedOrderItemTrashed(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find order items by Order ID
// @Tags Order Item Query
// @Description Retrieve all items belonging to a specific order
// @Accept json
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {object} response.ApiResponsesOrderItem "List of items in order"
// @Failure 400 {object} errors.ErrorResponse "Invalid Order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve order item data"
// @Router /api/order-item/order/{order_id} [get]
func (h *orderItemQueryHandlerApi) FindOrderItemByOrder(w http.ResponseWriter, r *http.Request) error {
	orderId, err := strconv.Atoi(chi.URLParam(r, "order_id"))
	if err != nil || orderId <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Order ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedOrderItems(ctx, orderId); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(orderId)})
	if err != nil {
		return h.handleGrpcError(err, "FindOrderItemByOrder")
	}

	apiResponse := h.mapper.ToApiResponsesOrderItem(res)
	h.cache.SetCachedOrderItems(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

func (h *orderItemQueryHandlerApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to " + operation)
	return sharedErrors.ParseGrpcError(err)
}
