package shippingaddresshandler

import (
	"net/http"
	"strconv"

	shippingaddress_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/shipping_address"
	pbshippingaddress "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type shippingAddressCommandHandleApi struct {
	client pbshippingaddress.ShippingCommandServiceClient
	logger logger.LoggerInterface
	mapper apimapper.ShippingAddressCommandResponseMapper
	cache  shippingaddress_cache.ShippingAddressCommandCache
}

type shippingAddressCommandHandleDeps struct {
	client pbshippingaddress.ShippingCommandServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.ShippingAddressCommandResponseMapper
	cache  shippingaddress_cache.ShippingAddressCommandCache
}

func NewShippingAddressCommandHandleApi(deps *shippingAddressCommandHandleDeps) {
	handler := &shippingAddressCommandHandleApi{
		client: deps.client,
		logger: deps.logger,
		mapper: deps.mapper,
		cache:  deps.cache,
	}

	deps.router.Route("/api/shipping-address-command", func(router chi.Router) {
		router.Post("/trashed/{id}", httpx.Handler(handler.TrashedShippingAddress))
		router.Post("/restore/{id}", httpx.Handler(handler.RestoreShippingAddress))
		router.Delete("/permanent/{id}", httpx.Handler(handler.DeleteShippingAddressPermanent))
		router.Post("/restore/all", httpx.Handler(handler.RestoreAllShippingAddress))
		router.Post("/permanent/all", httpx.Handler(handler.DeleteAllShippingAddressPermanent))
	})
}

// @Security Bearer
// @Summary Move shipping address to trash
// @Tags Shipping Address Command
// @Description Move a shipping address record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Shipping Address ID"
// @Success 200 {object} response.ApiResponseShippingAddressDeleteAt "Successfully moved shipping address to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid shipping address ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move shipping address to trash"
// @Router /api/shipping-address-command/trashed/{id} [post]
func (h *shippingAddressCommandHandleApi) TrashedShippingAddress(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedShipping(ctx, &pbshippingaddress.FindByIdShippingRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Trash")
	}

	h.cache.DeleteShippingAddressCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseShippingAddressDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed shipping address
// @Tags Shipping Address Command
// @Description Restore a trashed shipping address record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Shipping Address ID"
// @Success 200 {object} response.ApiResponseShippingAddressDeleteAt "Successfully restored shipping address"
// @Failure 400 {object} errors.ErrorResponse "Invalid shipping address ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore shipping address"
// @Router /api/shipping-address-command/restore/{id} [post]
func (h *shippingAddressCommandHandleApi) RestoreShippingAddress(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreShipping(ctx, &pbshippingaddress.FindByIdShippingRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Restore")
	}

	h.cache.DeleteShippingAddressCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseShippingAddressDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a shipping address
// @Tags Shipping Address Command
// @Description Permanently delete a shipping address record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Shipping Address ID"
// @Success 200 {object} response.ApiResponseShippingAddressDelete "Successfully deleted shipping address record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid shipping address ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete shipping address permanently"
// @Router /api/shipping-address-command/permanent/{id} [delete]
func (h *shippingAddressCommandHandleApi) DeleteShippingAddressPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteShippingPermanent(ctx, &pbshippingaddress.FindByIdShippingRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Delete")
	}

	h.cache.DeleteShippingAddressCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseShippingAddressDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed shipping addresses
// @Tags Shipping Address Command
// @Description Restore all trashed shipping address records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseShippingAddressAll "Successfully restored all shipping addresses"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore shipping addresses"
// @Router /api/shipping-address-command/restore/all [post]
func (h *shippingAddressCommandHandleApi) RestoreAllShippingAddress(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllShipping(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "RestoreAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseShippingAddressAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed shipping addresses
// @Tags Shipping Address Command
// @Description Permanently delete all trashed shipping address records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseShippingAddressAll "Successfully deleted all shipping addresses permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete shipping addresses permanently"
// @Router /api/shipping-address-command/permanent/all [post]
func (h *shippingAddressCommandHandleApi) DeleteAllShippingAddressPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllShippingPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "DeleteAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseShippingAddressAll(res))
}

func (h *shippingAddressCommandHandleApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to "+operation, zap.Error(err))
	return sharedErrors.ParseGrpcError(err)
}
