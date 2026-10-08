package orderhandler

import (
	pbshippingaddress "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
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
	"google.golang.org/protobuf/types/known/emptypb"
)

type orderCommandHandlerApi struct {
	client pborder.OrderCommandServiceClient
	logger logger.LoggerInterface
	mapper apimapper.OrderCommandResponseMapper
	cache  order_cache.OrderCommandCache
}

type orderCommandHandleDeps struct {
	client pborder.OrderCommandServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.OrderCommandResponseMapper
	cache  order_cache.OrderCommandCache
}

func NewOrderCommandHandleApi(params *orderCommandHandleDeps) *orderCommandHandlerApi {
	handler := &orderCommandHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/order-command", func(routerOrder chi.Router) {
		routerOrder.Post("/create", httpx.Handler(handler.Create))
		routerOrder.Post("/update/{id}", httpx.Handler(handler.Update))
		routerOrder.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerOrder.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerOrder.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerOrder.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerOrder.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new order
// @Tags Order Command
// @Description Create a new order with items and shipping address
// @Accept json
// @Produce json
// @Param request body requests.CreateOrderRequest true "Order details"
// @Success 200 {object} response.ApiResponseOrder "Successfully created order"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} errors.ErrorResponse "Failed to create order"
// @Router /api/order-command/create [post]
func (h *orderCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateOrderRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()

	items := make([]*pborder.CreateOrderItemRequest, 0)
	for _, item := range body.Items {
		items = append(items, &pborder.CreateOrderItemRequest{
			ProductId: int32(item.ProductID),
			Quantity:  int32(item.Quantity),
			Price:     int32(item.Price),
		})
	}

	res, err := h.client.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(body.MerchantID),
		UserId:     int32(body.UserID),
		TotalPrice: int32(body.TotalPrice),
		Items:      items,
		Shipping: &pbshippingaddress.CreateShippingAddressRequest{
			Alamat:         body.ShippingAddress.Alamat,
			Provinsi:       body.ShippingAddress.Provinsi,
			Kota:           body.ShippingAddress.Kota,
			Courier:        body.ShippingAddress.Courier,
			ShippingMethod: body.ShippingAddress.ShippingMethod,
			ShippingCost:   int32(body.ShippingAddress.ShippingCost),
			Negara:         body.ShippingAddress.Negara,
		},
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrder(res))
}

// @Security Bearer
// @Summary Update an existing order
// @Tags Order Command
// @Description Update an existing order record
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body requests.UpdateOrderRequest true "Updated order details"
// @Success 200 {object} response.ApiResponseOrder "Successfully updated order"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} errors.ErrorResponse "Failed to update order"
// @Router /api/order-command/update/{id} [post]
func pointerValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func (h *orderCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateOrderRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()

	items := make([]*pborder.UpdateOrderItemRequest, 0)
	for _, item := range body.Items {
		items = append(items, &pborder.UpdateOrderItemRequest{
			OrderItemId: int32(item.OrderItemID),
			ProductId:   int32(item.ProductID),
			Quantity:    int32(item.Quantity),
			Price:       int32(item.Price),
		})
	}

	updateRequest := &pborder.UpdateOrderRequest{
		OrderId:    int32(id),
		UserId:     int32(body.UserID),
		TotalPrice: int32(body.TotalPrice),
		Items:      items,
	}
	if shipping := body.ShippingAddress; shipping != nil {
		updateRequest.Shipping = &pbshippingaddress.UpdateShippingAddressRequest{
			ShippingId:     int32(pointerValue(shipping.ShippingID)),
			Alamat:         shipping.Alamat,
			Provinsi:       shipping.Provinsi,
			Kota:           shipping.Kota,
			Courier:        shipping.Courier,
			ShippingMethod: shipping.ShippingMethod,
			ShippingCost:   int32(shipping.ShippingCost),
			Negara:         shipping.Negara,
		}
	}

	res, err := h.client.Update(ctx, updateRequest)
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrder(res))
}

// @Security Bearer
// @Summary Move order to trash
// @Tags Order Command
// @Description Move an order record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ApiResponseOrderDeleteAt "Successfully moved order to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move order to trash"
// @Router /api/order-command/trashed/{id} [post]
func (h *orderCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed order
// @Tags Order Command
// @Description Restore a trashed order record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ApiResponseOrderDeleteAt "Successfully restored order"
// @Failure 400 {object} errors.ErrorResponse "Invalid order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore order"
// @Router /api/order-command/restore/{id} [post]
func (h *orderCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreOrder(ctx, &pborder.FindByIdOrderRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete an order
// @Tags Order Command
// @Description Permanently delete an order record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ApiResponseOrderDelete "Successfully deleted order record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete order permanently"
// @Router /api/order-command/permanent/{id} [delete]
func (h *orderCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteOrderPermanent(ctx, &pborder.FindByIdOrderRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed orders
// @Tags Order Command
// @Description Restore all trashed order records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseOrderAll "Successfully restored all orders"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore orders"
// @Router /api/order-command/restore/all [post]
func (h *orderCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllOrder(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed orders
// @Tags Order Command
// @Description Permanently delete all trashed order records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseOrderAll "Successfully deleted all orders permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete orders permanently"
// @Router /api/order-command/permanent/all [post]
func (h *orderCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllOrderPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteOrderCache(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseOrderAll(res))
}
