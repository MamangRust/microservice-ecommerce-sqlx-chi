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
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/banner"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbbanner "github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type bannerCommandHandlerApi struct {
	client        pbbanner.BannerCommandServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.BannerCommandResponseMapper
	cache         banner_cache.BannerCommandCache
	observability observability.TraceLoggerObservability
}

type bannerCommandHandleDeps struct {
	client        pbbanner.BannerCommandServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.BannerCommandResponseMapper
	cache         banner_cache.BannerCommandCache
	observability observability.TraceLoggerObservability
}

func NewBannerCommandHandleApi(params *bannerCommandHandleDeps) *bannerCommandHandlerApi {
	handler := &bannerCommandHandlerApi{
		client:        params.client,
		logger:        params.logger,
		mapper:        params.mapper,
		cache:         params.cache,
		observability: params.observability,
	}

	params.router.Route("/api/banner-command", func(routerBanner chi.Router) {
		routerBanner.Post("/create", httpx.Handler(handler.Create))
		routerBanner.Post("/update/{id}", httpx.Handler(handler.Update))
		routerBanner.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerBanner.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerBanner.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerBanner.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerBanner.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new banner
// @Tags Banner Command
// @Description Create a new banner
// @Accept json
// @Produce json
// @Param request body requests.CreateBannerRequest true "Banner details"
// @Success 200 {object} response.ApiResponseBanner "Banner created"
// @Failure 400 {object} errors.ErrorResponse "Invalid request"
// @Failure 500 {object} errors.ErrorResponse "Failed to create banner"
// @Router /api/banner-command/create [post]
func (h *bannerCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"CreateBanner",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	var body requests.CreateBannerRequest
	if err := httpx.Bind(r, &body); err != nil {
		status = "error"
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		status = "error"
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.client.Create(ctx, &pbbanner.CreateBannerRequest{
		Name:      body.Name,
		StartDate: body.StartDate,
		EndDate:   body.EndDate,
		StartTime: body.StartTime,
		EndTime:   body.EndTime,
		IsActive:  body.IsActive,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Create")
	}

	logSuccess("Banner created successfully")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBanner(res))
}

// @Security Bearer
// @Summary Update a banner
// @Tags Banner Command
// @Description Update an existing banner
// @Accept json
// @Produce json
// @Param id path int true "Banner ID"
// @Param request body requests.UpdateBannerRequest true "Banner details"
// @Success 200 {object} response.ApiResponseBanner "Banner updated"
// @Failure 400 {object} errors.ErrorResponse "Invalid request"
// @Failure 500 {object} errors.ErrorResponse "Failed to update banner"
// @Router /api/banner-command/update/{id} [post]
func (h *bannerCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"UpdateBanner",
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

	var body requests.UpdateBannerRequest
	if err := httpx.Bind(r, &body); err != nil {
		status = "error"
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		status = "error"
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := h.client.Update(ctx, &pbbanner.UpdateBannerRequest{
		BannerId:  int32(id),
		Name:      body.Name,
		StartDate: body.StartDate,
		EndDate:   body.EndDate,
		StartTime: body.StartTime,
		EndTime:   body.EndTime,
		IsActive:  body.IsActive,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Update")
	}

	h.cache.DeleteBannerCache(ctx, id)

	logSuccess("Banner updated successfully")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBanner(res))
}

// @Security Bearer
// @Summary Trash a banner
// @Tags Banner Command
// @Description Move a banner to trash
// @Accept json
// @Produce json
// @Param id path int true "Banner ID"
// @Success 200 {object} response.ApiResponseBannerDeleteAt "Banner trashed"
// @Failure 400 {object} errors.ErrorResponse "Invalid ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to trash banner"
// @Router /api/banner-command/trashed/{id} [post]
func (h *bannerCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"TrashBanner",
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

	res, err := h.client.Trash(ctx, &pbbanner.FindByIdBannerRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Trash")
	}

	h.cache.DeleteBannerCache(ctx, id)

	logSuccess("Banner moved to trash")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBannerDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a banner
// @Tags Banner Command
// @Description Restore a trashed banner
// @Accept json
// @Produce json
// @Param id path int true "Banner ID"
// @Success 200 {object} response.ApiResponseBannerDeleteAt "Banner restored"
// @Failure 400 {object} errors.ErrorResponse "Invalid ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore banner"
// @Router /api/banner-command/restore/{id} [post]
func (h *bannerCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"RestoreBanner",
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

	res, err := h.client.Restore(ctx, &pbbanner.FindByIdBannerRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Restore")
	}

	h.cache.DeleteBannerCache(ctx, id)

	logSuccess("Banner restored successfully")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBannerDeleteAt(res))
}

// @Security Bearer
// @Summary Delete a banner permanently
// @Tags Banner Command
// @Description Permanently delete a banner
// @Accept json
// @Produce json
// @Param id path int true "Banner ID"
// @Success 200 {object} response.ApiResponseBannerDelete "Banner deleted"
// @Failure 400 {object} errors.ErrorResponse "Invalid ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete banner"
// @Router /api/banner-command/permanent/{id} [delete]
func (h *bannerCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"DeletePermanent",
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

	res, err := h.client.DeletePermanent(ctx, &pbbanner.FindByIdBannerRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Delete")
	}

	h.cache.DeleteBannerCache(ctx, id)

	logSuccess("Banner deleted permanently")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBannerDelete(res))
}

// @Security Bearer
// @Summary Restore all banners
// @Tags Banner Command
// @Description Restore all trashed banners
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseBannerAll "All banners restored"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore banners"
// @Router /api/banner-command/restore/all [post]
func (h *bannerCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"RestoreAll",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	res, err := h.client.RestoreAll(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "RestoreAll")
	}

	logSuccess("All banners restored")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBannerAll(res))
}

// @Security Bearer
// @Summary Delete all banners permanently
// @Tags Banner Command
// @Description Permanently delete all banners
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseBannerAll "All banners deleted"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete banners"
// @Router /api/banner-command/permanent/all [post]
func (h *bannerCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(
		r.Context(),
		"DeleteAllPermanent",
		attribute.String("path", r.URL.Path),
		attribute.String("method", r.Method),
	)
	defer end(status)
	r = r.WithContext(ctx)

	res, err := h.client.DeleteAll(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "DeleteAll")
	}

	logSuccess("All banners deleted permanently")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseBannerAll(res))
}

func (h *bannerCommandHandlerApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		fmt.Sprintf("Banner command error in %s", method),
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
