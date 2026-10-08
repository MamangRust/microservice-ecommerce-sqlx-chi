package shippingaddresshandler

import (
	"net/http"
	"strconv"

	shippingaddress_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/shipping_address"
	pbshippingaddress "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type shippingAddressQueryHandleApi struct {
	client pbshippingaddress.ShippingQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.ShippingAddressQueryResponseMapper
	cache  shippingaddress_cache.ShippingAddressQueryCache
}

type shippingAddressQueryHandleDeps struct {
	client pbshippingaddress.ShippingQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.ShippingAddressQueryResponseMapper
	cache  shippingaddress_cache.ShippingAddressQueryCache
}

func NewShippingAddressQueryHandleApi(deps *shippingAddressQueryHandleDeps) {
	handler := &shippingAddressQueryHandleApi{
		client: deps.client,
		logger: deps.logger,
		mapper: deps.mapper,
		cache:  deps.cache,
	}

	deps.router.Route("/api/shipping-address-query", func(router chi.Router) {
		router.Get("/", httpx.Handler(handler.FindAll))
		router.Get("/{id}", httpx.Handler(handler.FindById))
		router.Get("/order/{id}", httpx.Handler(handler.FindByOrder))
		router.Get("/active", httpx.Handler(handler.FindByActive))
		router.Get("/trashed", httpx.Handler(handler.FindByTrashed))
	})
}

// @Security Bearer
// @Summary Find all shipping addresses
// @Tags Shipping Address Query
// @Description Retrieve a list of all shipping addresses
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationShippingAddress "List of shipping addresses"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve shipping address data"
// @Router /api/shipping-address-query [get]
func (h *shippingAddressQueryHandleApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllShippingAddress{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetShippingAddressAllCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbshippingaddress.FindAllShippingRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindAll(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindAll")
	}

	response := h.mapper.ToApiResponsePaginationShippingAddress(res)
	h.cache.SetShippingAddressAllCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Find shipping address by ID
// @Tags Shipping Address Query
// @Description Retrieve a shipping address by ID
// @Accept json
// @Produce json
// @Param id path int true "Shipping Address ID"
// @Success 200 {object} response.ApiResponseShippingAddress "Shipping address data"
// @Failure 400 {object} errors.ErrorResponse "Invalid shipping address ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve shipping address data"
// @Router /api/shipping-address-query/{id} [get]
func (h *shippingAddressQueryHandleApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cached, found := h.cache.GetCachedShippingAddressCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbshippingaddress.FindByIdShippingRequest{Id: int32(id)}
	res, err := h.client.FindById(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindById")
	}

	response := h.mapper.ToApiResponseShippingAddress(res)
	h.cache.SetCachedShippingAddressCache(ctx, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Find shipping address by Order ID
// @Tags Shipping Address Query
// @Description Retrieve a shipping address for a specific order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.ApiResponseShippingAddress "Shipping address data"
// @Failure 400 {object} errors.ErrorResponse "Invalid order ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve shipping address data"
// @Router /api/shipping-address-query/order/{id} [get]
func (h *shippingAddressQueryHandleApi) FindByOrder(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cached, found := h.cache.GetCachedShippingAddressByOrderCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbshippingaddress.FindByIdShippingRequest{Id: int32(id)}
	res, err := h.client.FindByOrder(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindByOrder")
	}

	response := h.mapper.ToApiResponseShippingAddress(res)
	h.cache.SetCachedShippingAddressByOrderCache(ctx, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve active shipping addresses
// @Tags Shipping Address Query
// @Description Retrieve a list of active shipping addresses
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationShippingAddressDeleteAt "List of active shipping addresses"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve shipping address data"
// @Router /api/shipping-address-query/active [get]
func (h *shippingAddressQueryHandleApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllShippingAddress{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetShippingAddressActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbshippingaddress.FindAllShippingRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByActive(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindByActive")
	}

	response := h.mapper.ToApiResponsePaginationShippingAddressDeleteAt(res)
	h.cache.SetShippingAddressActiveCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve trashed shipping addresses
// @Tags Shipping Address Query
// @Description Retrieve a list of trashed shipping address records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationShippingAddressDeleteAt "List of trashed shipping address data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve shipping address data"
// @Router /api/shipping-address-query/trashed [get]
func (h *shippingAddressQueryHandleApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllShippingAddress{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetShippingAddressTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbshippingaddress.FindAllShippingRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByTrashed(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindByTrashed")
	}

	response := h.mapper.ToApiResponsePaginationShippingAddressDeleteAt(res)
	h.cache.SetShippingAddressTrashedCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

func (h *shippingAddressQueryHandleApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to "+operation, zap.Error(err))
	return sharedErrors.ParseGrpcError(err)
}
