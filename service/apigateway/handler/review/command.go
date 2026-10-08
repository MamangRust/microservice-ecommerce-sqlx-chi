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
	"google.golang.org/protobuf/types/known/emptypb"
)

type reviewCommandHandleApi struct {
	client        pbreview.ReviewCommandServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewCommandResponseMapper
	queryMapper   apimapper.ReviewQueryResponseMapper
	cache         review_cache.ReviewCommandCache
	observability observability.TraceLoggerObservability
}

type reviewCommandHandleDeps struct {
	client        pbreview.ReviewCommandServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewCommandResponseMapper
	queryMapper   apimapper.ReviewQueryResponseMapper
	cache         review_cache.ReviewCommandCache
	observability observability.TraceLoggerObservability
}

func NewReviewCommandHandleApi(deps *reviewCommandHandleDeps) {
	handler := &reviewCommandHandleApi{
		client:        deps.client,
		logger:        deps.logger,
		mapper:        deps.mapper,
		queryMapper:   deps.queryMapper,
		cache:         deps.cache,
		observability: deps.observability,
	}

	deps.router.Route("/api/review-command", func(router chi.Router) {
		router.Post("/create", httpx.Handler(handler.Create))
		router.Post("/update/{id}", httpx.Handler(handler.Update))
		router.Post("/trashed/{id}", httpx.Handler(handler.TrashedReview))
		router.Post("/restore/{id}", httpx.Handler(handler.RestoreReview))
		router.Delete("/permanent/{id}", httpx.Handler(handler.DeleteReviewPermanent))
		router.Post("/restore/all", httpx.Handler(handler.RestoreAllReview))
		router.Post("/permanent/all", httpx.Handler(handler.DeleteAllReviewPermanent))
	})
}

// @Security Bearer
// @Summary Create a new review
// @Tags Review Command
// @Description Create a new product review for the authenticated user
// @Accept json
// @Produce json
// @Param body body requests.CreateReviewRequest true "Create review request"
// @Success 201 {object} response.ApiResponseReview "Successfully created review"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create review"
// @Router /api/review-command/create [post]
func (h *reviewCommandHandleApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateReviewRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok {
		return httpx.NewHTTPError(http.StatusUnauthorized, "User ID not found")
	}

	ctx := r.Context()
	grpcReq := &pbreview.CreateReviewRequest{
		UserId:    int32(userID),
		ProductId: int32(body.ProductID),
		Comment:   body.Comment,
		Rating:    int32(body.Rating),
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "CreateReview")
	defer func() {
		end(status)
	}()

	res, err := h.client.Create(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Create")
	}

	logSuccess("Successfully created review")
	return httpx.JSON(w, http.StatusCreated, h.queryMapper.ToApiResponseReview(res))
}

// @Security Bearer
// @Summary Update review
// @Tags Review Command
// @Description Update an existing review's comment and rating
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Param body body requests.UpdateReviewRequest true "Update review request"
// @Success 200 {object} response.ApiResponseReview "Successfully updated review"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update review"
// @Router /api/review-command/update/{id} [post]
func (h *reviewCommandHandleApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateReviewRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	grpcReq := &pbreview.UpdateReviewRequest{
		ReviewId: int32(id),
		Comment:  body.Comment,
		Rating:   int32(body.Rating),
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "UpdateReview")
	defer func() {
		end(status)
	}()

	res, err := h.client.Update(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Update")
	}

	h.cache.DeleteReviewCache(ctx, id)

	logSuccess("Successfully updated review")
	return httpx.JSON(w, http.StatusOK, h.queryMapper.ToApiResponseReview(res))
}

// @Security Bearer
// @Summary Move review to trash
// @Tags Review Command
// @Description Move a review record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} response.ApiResponseReviewDeleteAt "Successfully moved review to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid review ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move review to trash"
// @Router /api/review-command/trashed/{id} [post]
func (h *reviewCommandHandleApi) TrashedReview(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "TrashedReview")
	defer func() {
		end(status)
	}()

	res, err := h.client.TrashedReview(ctx, &pbreview.FindByIdReviewRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Trash")
	}

	h.cache.DeleteReviewCache(ctx, id)

	logSuccess("Successfully moved review to trash")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewDeleteAt(res))
}

// @Security Bearer
// @Summary Restore trashed review
// @Tags Review Command
// @Description Restore a trashed review record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} response.ApiResponseReviewDeleteAt "Successfully restored review"
// @Failure 400 {object} errors.ErrorResponse "Invalid review ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore review"
// @Router /api/review-command/restore/{id} [post]
func (h *reviewCommandHandleApi) RestoreReview(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreReview")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreReview(ctx, &pbreview.FindByIdReviewRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Restore")
	}

	h.cache.DeleteReviewCache(ctx, id)

	logSuccess("Successfully restored review")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete review
// @Tags Review Command
// @Description Permanently delete a review record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review ID"
// @Success 200 {object} response.ApiResponseReviewDelete "Successfully deleted review record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid review ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete review permanently"
// @Router /api/review-command/permanent/{id} [delete]
func (h *reviewCommandHandleApi) DeleteReviewPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteReviewPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteReviewPermanent(ctx, &pbreview.FindByIdReviewRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Delete")
	}

	h.cache.DeleteReviewCache(ctx, id)

	logSuccess("Successfully deleted review permanently")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed reviews
// @Tags Review Command
// @Description Restore all trashed review records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseReviewAll "Successfully restored all reviews"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore reviews"
// @Router /api/review-command/restore/all [post]
func (h *reviewCommandHandleApi) RestoreAllReview(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreAllReview")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreAllReview(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "RestoreAll")
	}

	logSuccess("Successfully restored all reviews")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed reviews
// @Tags Review Command
// @Description Permanently delete all trashed review records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseReviewAll "Successfully deleted all reviews permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete reviews permanently"
// @Router /api/review-command/permanent/all [post]
func (h *reviewCommandHandleApi) DeleteAllReviewPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteAllReviewPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteAllReviewPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "DeleteAll")
	}

	logSuccess("Successfully deleted all reviews permanently")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewAll(res))
}

func (h *reviewCommandHandleApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		"Review command error in "+method,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
