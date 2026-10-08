package reviewhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	review_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbreview "github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type reviewQueryHandleApi struct {
	client        pbreview.ReviewQueryServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewQueryResponseMapper
	cache         review_cache.ReviewQueryCache
	observability observability.TraceLoggerObservability
}

type reviewQueryHandleDeps struct {
	client        pbreview.ReviewQueryServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewQueryResponseMapper
	cache         review_cache.ReviewQueryCache
	observability observability.TraceLoggerObservability
}

func NewReviewQueryHandleApi(deps *reviewQueryHandleDeps) {
	handler := &reviewQueryHandleApi{
		client:        deps.client,
		logger:        deps.logger,
		mapper:        deps.mapper,
		cache:         deps.cache,
		observability: deps.observability,
	}

	deps.router.Route("/api/review-query", func(router chi.Router) {
		router.Get("/", httpx.Handler(handler.FindAll))
		router.Get("/product/{id}", httpx.Handler(handler.FindByProduct))
		router.Get("/merchant/{id}", httpx.Handler(handler.FindByMerchant))
		router.Get("/active", httpx.Handler(handler.FindByActive))
		router.Get("/trashed", httpx.Handler(handler.FindByTrashed))
	})
}

// @Security Bearer
// @Summary Find all reviews
// @Tags Review Query
// @Description Retrieve a list of all reviews
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationReview "List of reviews"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve review data"
// @Router /api/review-query [get]
func (h *reviewQueryHandleApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetReviewAllCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbreview.FindAllReviewRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindAllReviews")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindAll(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindAll")
	}

	response := h.mapper.ToApiResponsePaginationReview(res)
	h.cache.SetReviewAllCache(ctx, req, response)

	logSuccess("Successfully retrieved reviews")
	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Find reviews by Product ID
// @Tags Review Query
// @Description Retrieve reviews for a specific product including review details
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationReviewsDetail "List of reviews with details"
// @Failure 400 {object} errors.ErrorResponse "Invalid Product ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve review data"
// @Router /api/review-query/product/{id} [get]
func (h *reviewQueryHandleApi) FindByProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Product ID")
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	grpcReq := &pbreview.FindAllReviewProductRequest{
		ProductId: int32(id),
		Page:      int32(page),
		PageSize:  int32(pageSize),
		Search:    search,
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByProductReviews")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByProduct(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByProduct")
	}

	response := h.mapper.ToApiResponsePaginationReviewsDetail(res)

	logSuccess("Successfully retrieved reviews for product")
	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Find reviews by Merchant ID
// @Tags Review Query
// @Description Retrieve reviews for a specific merchant including review details
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationReviewsDetail "List of reviews with details"
// @Failure 400 {object} errors.ErrorResponse "Invalid Merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve review data"
// @Router /api/review-query/merchant/{id} [get]
func (h *reviewQueryHandleApi) FindByMerchant(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Merchant ID")
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllReviewByMerchant{
		MerchantID: id,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetReviewByMerchantCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbreview.FindAllReviewMerchantRequest{
		MerchantId: int32(id),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		Search:     search,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByMerchantReviews")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByMerchant(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByMerchant")
	}

	response := h.mapper.ToApiResponsePaginationReviewsDetail(res)
	h.cache.SetReviewByMerchantCache(ctx, req, response)

	logSuccess("Successfully retrieved reviews for merchant")
	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve active reviews
// @Tags Review Query
// @Description Retrieve a list of active reviews
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationReviewDeleteAt "List of active reviews"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve review data"
// @Router /api/review-query/active [get]
func (h *reviewQueryHandleApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetReviewActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbreview.FindAllReviewRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByActiveReviews")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByActive(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByActive")
	}

	response := h.mapper.ToApiResponsePaginationReviewDeleteAt(res)
	h.cache.SetReviewActiveCache(ctx, req, response)

	logSuccess("Successfully retrieved active reviews")
	return httpx.JSON(w, http.StatusOK, response)
}

// @Security Bearer
// @Summary Retrieve trashed reviews
// @Tags Review Query
// @Description Retrieve a list of trashed review records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationReviewDeleteAt "List of trashed reviews"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve review data"
// @Router /api/review-query/trashed [get]
func (h *reviewQueryHandleApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	req := &requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	ctx := r.Context()
	if cached, found := h.cache.GetReviewTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cached)
	}

	grpcReq := &pbreview.FindAllReviewRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByTrashedReviews")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByTrashed(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByTrashed")
	}

	response := h.mapper.ToApiResponsePaginationReviewDeleteAt(res)
	h.cache.SetReviewTrashedCache(ctx, req, response)

	logSuccess("Successfully retrieved trashed reviews")
	return httpx.JSON(w, http.StatusOK, response)
}

func (h *reviewQueryHandleApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		"Review query error in "+method,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
