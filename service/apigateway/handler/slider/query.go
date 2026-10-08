package sliderhandler

import (
	"net/http"
	"strconv"

	slider_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/slider"
	pbslider "github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type sliderQueryHandleApi struct {
	client pbslider.SliderQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.SliderQueryResponseMapper
	cache  slider_cache.SliderQueryCache
}

type sliderQueryHandleDeps struct {
	client pbslider.SliderQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.SliderQueryResponseMapper
	cache  slider_cache.SliderQueryCache
}

func NewSliderQueryHandleApi(deps *sliderQueryHandleDeps) {
	handler := &sliderQueryHandleApi{
		client: deps.client,
		logger: deps.logger,
		mapper: deps.mapper,
		cache:  deps.cache,
	}

	deps.router.Route("/api/slider-query", func(router chi.Router) {
		router.Get("/", httpx.Handler(handler.FindAll))
		router.Get("/{id}", httpx.Handler(handler.FindById))
		router.Get("/active", httpx.Handler(handler.FindByActive))
		router.Get("/trashed", httpx.Handler(handler.FindByTrashed))
	})
}

// @Security Bearer
// @Summary Find all sliders
// @Tags Slider Query
// @Description Retrieve a list of all sliders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationSlider "List of sliders"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve slider data"
// @Router /api/slider-query [get]
func (h *sliderQueryHandleApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetSliderAllCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbslider.FindAllSliderRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindAll(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindAll")
	}

	response := h.mapper.ToApiResponsePaginationSlider(res)
	h.cache.SetSliderAllCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Find slider by ID
// @Tags Slider Query
// @Description Retrieve a slider by ID
// @Accept json
// @Produce json
// @Param id path int true "Slider ID"
// @Success 200 {object} response.ApiResponseSlider "Slider data"
// @Failure 400 {object} errors.ErrorResponse "Invalid slider ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve slider data"
// @Router /api/slider-query/{id} [get]
func (h *sliderQueryHandleApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	if cached, found := h.cache.GetCachedSliderCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	res, err := h.client.FindById(ctx, &pbslider.FindByIdSliderRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "FindById")
	}

	response := h.mapper.ToApiResponseSlider(res)
	h.cache.SetCachedSliderCache(ctx, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve active sliders
// @Tags Slider Query
// @Description Retrieve a list of active sliders
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationSliderDeleteAt "List of active sliders"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve slider data"
// @Router /api/slider-query/active [get]
func (h *sliderQueryHandleApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetSliderActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbslider.FindAllSliderRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByActive(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindByActive")
	}

	response := h.mapper.ToApiResponsePaginationSliderDeleteAt(res)
	h.cache.SetSliderActiveCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve trashed sliders
// @Tags Slider Query
// @Description Retrieve a list of trashed slider records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationSliderDeleteAt "List of trashed sliders"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve slider data"
// @Router /api/slider-query/trashed [get]
func (h *sliderQueryHandleApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetSliderTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbslider.FindAllSliderRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.client.FindByTrashed(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "FindByTrashed")
	}

	response := h.mapper.ToApiResponsePaginationSliderDeleteAt(res)
	h.cache.SetSliderTrashedCache(ctx, req, response)

	return httpx.JSON(w, http.StatusOK, response)
}

func (h *sliderQueryHandleApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to "+operation, zap.Error(err))
	return sharedErrors.ParseGrpcError(err)
}
