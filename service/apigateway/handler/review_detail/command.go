package reviewdetailhandler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	reviewdetail_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/review_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	reviewapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review_detail"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbreviewdetail "github.com/MamangRust/microservice-ecommerce-grpc-pb/review_detail"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type reviewDetailCommandHandleApi struct {
	client        pbreviewdetail.ReviewDetailCommandServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewDetailCommandResponseMapper
	queryMapper   apimapper.ReviewDetailQueryResponseMapper
	reviewMapper  reviewapimapper.ReviewCommandResponseMapper
	cache         reviewdetail_cache.ReviewDetailCommandCache
	upload        upload_image.ImageUploads
	observability observability.TraceLoggerObservability
}

type reviewDetailCommandHandleDeps struct {
	client        pbreviewdetail.ReviewDetailCommandServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.ReviewDetailCommandResponseMapper
	queryMapper   apimapper.ReviewDetailQueryResponseMapper
	reviewMapper  reviewapimapper.ReviewCommandResponseMapper
	cache         reviewdetail_cache.ReviewDetailCommandCache
	upload        upload_image.ImageUploads
	observability observability.TraceLoggerObservability
}

func NewReviewDetailCommandHandleApi(deps *reviewDetailCommandHandleDeps) {
	handler := &reviewDetailCommandHandleApi{
		client:        deps.client,
		logger:        deps.logger,
		mapper:        deps.mapper,
		queryMapper:   deps.queryMapper,
		reviewMapper:  deps.reviewMapper,
		cache:         deps.cache,
		upload:        deps.upload,
		observability: deps.observability,
	}

	deps.router.Route("/api/review-detail-command", func(router chi.Router) {
		router.Post("/create", httpx.Handler(handler.Create))
		router.Post("/update/{id}", httpx.Handler(handler.Update))
		router.Post("/trashed/{id}", httpx.Handler(handler.TrashedReviewDetail))
		router.Post("/restore/{id}", httpx.Handler(handler.RestoreReviewDetail))
		router.Delete("/permanent/{id}", httpx.Handler(handler.DeleteReviewDetailPermanent))
		router.Post("/restore/all", httpx.Handler(handler.RestoreAllReviewDetail))
		router.Post("/permanent/all", httpx.Handler(handler.DeleteAllReviewDetailPermanent))
	})
}

// @Security Bearer
// @Summary Create a new review detail
// @Tags Review Detail Command
// @Description Create a new review detail (e.g., photo/video attachment)
// @Accept mpfd
// @Produce json
// @Param review_id formData int true "Review ID"
// @Param type formData string true "Attachment type (e.g., image, video)"
// @Param caption formData string true "Attachment caption"
// @Param url formData file true "Attachment file"
// @Success 201 {object} response.ApiResponseReviewDetail "Successfully created review detail"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create review detail"
// @Router /api/review-detail-command/create [post]
func (h *reviewDetailCommandHandleApi) Create(w http.ResponseWriter, r *http.Request) error {
	formData, err := h.parseReviewDetailForm(w, r)
	if err != nil {
		return err
	}

	ctx := r.Context()
	grpcReq := &pbreviewdetail.CreateReviewDetailRequest{
		ReviewId: int32(formData.ReviewID),
		Type:     formData.Type,
		Url:      formData.Url,
		Caption:  formData.Caption,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "CreateReviewDetail")
	defer func() {
		end(status)
	}()

	res, err := h.client.Create(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Create")
	}

	logSuccess("Successfully created review detail")
	return httpx.JSON(w, http.StatusCreated, h.queryMapper.ToApiResponseReviewDetail(res))
}

// @Security Bearer
// @Summary Update review detail
// @Tags Review Detail Command
// @Description Update an existing review detail
// @Accept mpfd
// @Produce json
// @Param id path int true "Review Detail ID"
// @Param type formData string true "Attachment type"
// @Param caption formData string true "Attachment caption"
// @Param url formData file false "New attachment file"
// @Success 200 {object} response.ApiResponseReviewDetail "Successfully updated review detail"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update review detail"
// @Router /api/review-detail-command/update/{id} [post]
func (h *reviewDetailCommandHandleApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	formData, err := h.parseReviewDetailForm(w, r)
	if err != nil {
		return err
	}

	ctx := r.Context()
	grpcReq := &pbreviewdetail.UpdateReviewDetailRequest{
		ReviewDetailId: int32(id),
		Type:           formData.Type,
		Url:            formData.Url,
		Caption:        formData.Caption,
	}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "UpdateReviewDetail")
	defer func() {
		end(status)
	}()

	res, err := h.client.Update(ctx, grpcReq)
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Update")
	}

	h.cache.DeleteReviewDetailCache(ctx, id)

	logSuccess("Successfully updated review detail")
	return httpx.JSON(w, http.StatusOK, h.queryMapper.ToApiResponseReviewDetail(res))
}

// @Security Bearer
// @Summary Move review detail to trash
// @Tags Review Detail Command
// @Description Move a review detail record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review Detail ID"
// @Success 200 {object} response.ApiResponseReviewDetailDeleteAt "Successfully moved review detail to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid review detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move review detail to trash"
// @Router /api/review-detail-command/trashed/{id} [post]
func (h *reviewDetailCommandHandleApi) TrashedReviewDetail(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "TrashedReviewDetail")
	defer func() {
		end(status)
	}()

	res, err := h.client.TrashedReviewDetail(ctx, &pbreviewdetail.FindByIdReviewDetailRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Trash")
	}

	h.cache.DeleteReviewDetailCache(ctx, id)

	logSuccess("Successfully moved review detail to trash")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewDetailDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed review detail
// @Tags Review Detail Command
// @Description Restore a trashed review detail record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review Detail ID"
// @Success 200 {object} response.ApiResponseReviewDetailDeleteAt "Successfully restored review detail"
// @Failure 400 {object} errors.ErrorResponse "Invalid review detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore review detail"
// @Router /api/review-detail-command/restore/{id} [post]
func (h *reviewDetailCommandHandleApi) RestoreReviewDetail(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreReviewDetail")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreReviewDetail(ctx, &pbreviewdetail.FindByIdReviewDetailRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Restore")
	}

	h.cache.DeleteReviewDetailCache(ctx, id)

	logSuccess("Successfully restored review detail")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseReviewDetailDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a review detail
// @Tags Review Detail Command
// @Description Permanently delete a review detail record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Review Detail ID"
// @Success 200 {object} response.ApiResponseReviewDelete "Successfully deleted review detail record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid review detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete review detail permanently"
// @Router /api/review-detail-command/permanent/{id} [delete]
func (h *reviewDetailCommandHandleApi) DeleteReviewDetailPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteReviewDetailPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteReviewDetailPermanent(ctx, &pbreviewdetail.FindByIdReviewDetailRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Delete")
	}

	h.cache.DeleteReviewDetailCache(ctx, id)

	logSuccess("Successfully deleted review detail permanently")
	return httpx.JSON(w, http.StatusOK, h.reviewMapper.ToApiResponseReviewDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed review details
// @Tags Review Detail Command
// @Description Restore all trashed review detail records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseReviewAll "Successfully restored all review details"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore review details"
// @Router /api/review-detail-command/restore/all [post]
func (h *reviewDetailCommandHandleApi) RestoreAllReviewDetail(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreAllReviewDetail")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreAllReviewDetail(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "RestoreAll")
	}

	logSuccess("Successfully restored all review details")
	return httpx.JSON(w, http.StatusOK, h.reviewMapper.ToApiResponseReviewAll(res))
}

// @Security Bearer
// @Summary Permanently delete all review details
// @Tags Review Detail Command
// @Description Permanently delete all trashed review detail records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseReviewAll "Successfully deleted all review details permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete review details permanently"
// @Router /api/review-detail-command/permanent/all [post]
func (h *reviewDetailCommandHandleApi) DeleteAllReviewDetailPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteAllReviewDetailPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteAllReviewDetailPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "DeleteAll")
	}

	logSuccess("Successfully deleted all review details permanently")
	return httpx.JSON(w, http.StatusOK, h.reviewMapper.ToApiResponseReviewAll(res))
}

func (h *reviewDetailCommandHandleApi) parseReviewDetailForm(w http.ResponseWriter, r *http.Request) (requests.ReviewDetailFormData, error) {
	var formData requests.ReviewDetailFormData
	var err error

	reviewIDStr := r.FormValue("review_id")
	if reviewIDStr != "" {
		formData.ReviewID, err = strconv.Atoi(reviewIDStr)
		if err != nil || formData.ReviewID <= 0 {
			return formData, httpx.NewHTTPError(http.StatusBadRequest, "Invalid Review ID")
		}
	}

	formData.Type = strings.TrimSpace(r.FormValue("type"))
	if formData.Type == "" {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Type is required")
	}

	formData.Caption = strings.TrimSpace(r.FormValue("caption"))
	if formData.Caption == "" {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Caption is required")
	}

	_, file, err := r.FormFile("url")
	if err == nil {
		uploadPath, err := h.upload.ProcessImageUpload(w, "uploads/review_detail", file, false)
		if err != nil {
			return formData, err
		}
		formData.Url = uploadPath
	} else if r.FormValue("url") != "" {
		formData.Url = r.FormValue("url")
	}

	return formData, nil
}

func (h *reviewDetailCommandHandleApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		"Review detail command error in "+method,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
