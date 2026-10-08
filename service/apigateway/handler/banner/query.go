package bannerhandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	banner_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/banner"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/banner"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbbanner "github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/go-chi/chi/v5"

	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type bannerQueryHandlerApi struct {
	client        pbbanner.BannerQueryServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.BannerQueryResponseMapper
	cache         banner_cache.BannerQueryCache
	observability observability.TraceLoggerObservability
}

type bannerQueryHandleDeps struct {
	client        pbbanner.BannerQueryServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.BannerQueryResponseMapper
	cache         banner_cache.BannerQueryCache
	observability observability.TraceLoggerObservability
}

func NewBannerQueryHandleApi(params *bannerQueryHandleDeps) *bannerQueryHandlerApi {
	handler := &bannerQueryHandlerApi{
		client:        params.client,
		logger:        params.logger,
		mapper:        params.mapper,
		cache:         params.cache,
		observability: params.observability,
	}

	params.router.Route("/api/banner-query", func(routerBanner chi.Router) {
		routerBanner.Get("/", httpx.Handler(handler.FindAll))
		routerBanner.Get("/{id}", httpx.Handler(handler.FindById))
		routerBanner.Get("/active", httpx.Handler(handler.FindByActive))
		routerBanner.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all banners
// @Tags Banner Query
// @Description Retrieve a list of all banners
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationBanner "List of banners"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve banner data"
// @Router /api/banner-query [get]
func (h *bannerQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"FindAll",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllBanner{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedBanners(ctx, req); found {
		logSuccess("Serving from cache")
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbbanner.FindAllBannerRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindAll")
	}

	apiResponse := h.mapper.ToApiResponsePaginationBanner(res)
	h.cache.SetCachedBanners(ctx, req, apiResponse)

	logSuccess("Request completed successfully")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find a banner by ID
// @Tags Banner Query
// @Description Retrieve a single banner by its ID
// @Accept json
// @Produce json
// @Param id path int true "Banner ID"
// @Success 200 {object} response.ApiResponseBanner "Banner data"
// @Failure 400 {object} errors.ErrorResponse "Invalid ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve banner data"
// @Router /api/banner-query/{id} [get]
func (h *bannerQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"FindById",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		status = "error"
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	if cachedData, found := h.cache.GetCachedBanner(ctx, id); found {
		logSuccess("Serving from cache")
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindById(ctx, &pbbanner.FindByIdBannerRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindById")
	}

	apiResponse := h.mapper.ToApiResponseBanner(res)
	h.cache.SetCachedBanner(ctx, apiResponse)

	logSuccess("Request completed successfully")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find active banners
// @Tags Banner Query
// @Description Retrieve a list of active banners
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationBannerDeleteAt "List of active banners"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve banner data"
// @Router /api/banner-query/active [get]
func (h *bannerQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"FindByActive",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllBanner{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedActiveBanners(ctx, req); found {
		logSuccess("Serving from cache")
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByActive(ctx, &pbbanner.FindAllBannerRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByActive")
	}

	apiResponse := h.mapper.ToApiResponsePaginationBannerDeleteAt(res)
	h.cache.SetCachedActiveBanners(ctx, req, apiResponse)

	logSuccess("Request completed successfully")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find trashed banners
// @Tags Banner Query
// @Description Retrieve a list of trashed banners
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationBannerDeleteAt "List of trashed banners"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve banner data"
// @Router /api/banner-query/trashed [get]
func (h *bannerQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"FindByTrashed",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllBanner{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedTrashedBanners(ctx, req); found {
		logSuccess("Serving from cache")
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindByTrashed(ctx, &pbbanner.FindAllBannerRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByTrashed")
	}

	apiResponse := h.mapper.ToApiResponsePaginationBannerDeleteAt(res)
	h.cache.SetCachedTrashedBanners(ctx, req, apiResponse)

	logSuccess("Request completed successfully")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

func (h *bannerQueryHandlerApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		fmt.Sprintf("Banner query error in %s", method),
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
