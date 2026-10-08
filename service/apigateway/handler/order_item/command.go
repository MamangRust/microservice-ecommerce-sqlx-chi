package orderitemhandler

import (
	"net/http"
	"strconv"

	orderitem_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/order_item"
	pborderitem "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type orderItemCommandHandlerApi struct {
	client pborderitem.OrderItemCommandServiceClient
	logger logger.LoggerInterface
	mapper apimapper.OrderItemCommandResponseMapper
	cache  orderitem_cache.OrderItemCommandCache
}

type orderItemCommandHandleDeps struct {
	client pborderitem.OrderItemCommandServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.OrderItemCommandResponseMapper
	cache  orderitem_cache.OrderItemCommandCache
}

func NewOrderItemCommandHandleApi(params *orderItemCommandHandleDeps) *orderItemCommandHandlerApi {
	handler := &orderItemCommandHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	// Registered on the parent router (like the query handlers in query.go):
	// chi panics when the same path is mounted twice, and a mounted subrouter
	// here would shadow the query routes under /api/order-item.
	routerOrderItem := params.router
	routerOrderItem.Post("/api/order-item/trash/{id}", httpx.Handler(handler.Trash))
	routerOrderItem.Post("/api/order-item/restore/{id}", httpx.Handler(handler.Restore))
	routerOrderItem.Delete("/api/order-item/permanent/{id}", httpx.Handler(handler.DeletePermanent))
	routerOrderItem.Post("/api/order-item/restore/all", httpx.Handler(handler.RestoreAll))
	routerOrderItem.Post("/api/order-item/permanent/all", httpx.Handler(handler.DeleteAllPermanent))
	return handler
}

// @Security Bearer
// @Summary Move order item to trash
// @Tags Order Item Command
// @Description Move an order item record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} response.ApiResponseOrderItem "Successfully moved order item to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid item ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move order item to trash"
// @Router /api/order-item/trash/{id} [post]
func (h *orderItemCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashOrderItem(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Trash")
	}

	// We don't have the order_id here, so we might need to invalidate all or just use the ID if cache supports it
	h.cache.DeleteCachedOrderItemByOrderId(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderItem(res))
}

// @Security Bearer
// @Summary Restore trashed order item
// @Tags Order Item Command
// @Description Restore a trashed order item record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} response.ApiResponseOrderItem "Successfully restored order item"
// @Failure 400 {object} errors.ErrorResponse "Invalid item ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore order item"
// @Router /api/order-item/restore/{id} [post]
func (h *orderItemCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreOrderItem(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Restore")
	}

	h.cache.DeleteCachedOrderItemByOrderId(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderItem(res))
}

// @Security Bearer
// @Summary Permanently delete order item
// @Tags Order Item Command
// @Description Permanently delete an order item record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} response.ApiResponseOrderItemDelete "Successfully deleted order item record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid item ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete order item permanently"
// @Router /api/order-item/permanent/{id} [delete]
func (h *orderItemCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteOrderItemPermanent(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Delete")
	}

	h.cache.DeleteCachedOrderItemByOrderId(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderItemDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed order items
// @Tags Order Item Command
// @Description Restore all trashed order item records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseOrderItemAll "Successfully restored all order items"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore order items"
// @Router /api/order-item/restore/all [post]
func (h *orderItemCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllOrdersItem(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "RestoreAll")
	}

	h.cache.DeleteCachedOrderItemByOrderId(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderItemAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed order items
// @Tags Order Item Command
// @Description Permanently delete all trashed order item records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseOrderItemAll "Successfully deleted all order items permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete order items permanently"
// @Router /api/order-item/permanent/all [post]
func (h *orderItemCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllPermanentOrdersItem(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "DeleteAll")
	}

	h.cache.DeleteCachedOrderItemByOrderId(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderItemAll(res))
}

func (h *orderItemCommandHandlerApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to " + operation)
	return sharedErrors.ParseGrpcError(err)
}
