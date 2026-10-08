package sliderhandler

import (
	"net/http"
	"strconv"
	"strings"

	slider_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/slider"
	pbslider "github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type sliderCommandHandleApi struct {
	client      pbslider.SliderCommandServiceClient
	logger      logger.LoggerInterface
	mapper      apimapper.SliderCommandResponseMapper
	queryMapper apimapper.SliderQueryResponseMapper
	cache       slider_cache.SliderCommandCache
	upload      upload_image.ImageUploads
}

type sliderCommandHandleDeps struct {
	client      pbslider.SliderCommandServiceClient
	router      chi.Router
	logger      logger.LoggerInterface
	mapper      apimapper.SliderCommandResponseMapper
	queryMapper apimapper.SliderQueryResponseMapper
	cache       slider_cache.SliderCommandCache
	upload      upload_image.ImageUploads
}

func NewSliderCommandHandleApi(deps *sliderCommandHandleDeps) {
	handler := &sliderCommandHandleApi{
		client:      deps.client,
		logger:      deps.logger,
		mapper:      deps.mapper,
		queryMapper: deps.queryMapper,
		cache:       deps.cache,
		upload:      deps.upload,
	}

	deps.router.Route("/api/slider-command", func(router chi.Router) {
		router.Post("/create", httpx.Handler(handler.Create))
		router.Post("/update/{id}", httpx.Handler(handler.Update))
		router.Post("/trashed/{id}", httpx.Handler(handler.TrashedSlider))
		router.Post("/restore/{id}", httpx.Handler(handler.RestoreSlider))
		router.Delete("/permanent/{id}", httpx.Handler(handler.DeleteSliderPermanent))
		router.Post("/restore/all", httpx.Handler(handler.RestoreAllSlider))
		router.Post("/permanent/all", httpx.Handler(handler.DeleteAllSliderPermanent))
	})
}

// @Security Bearer
// @Summary Create slider
// @Tags Slider Command
// @Description Create a new slider with an image upload
// @Accept mpfd
// @Produce json
// @Param name formData string true "Slider name"
// @Param image_slider formData file true "Slider image"
// @Success 201 {object} response.ApiResponseSlider "Successfully created slider"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create slider"
// @Router /api/slider-command/create [post]
func (h *sliderCommandHandleApi) Create(w http.ResponseWriter, r *http.Request) error {
	formData, err := h.parseSliderForm(w, r, true)
	if err != nil {
		return err
	}

	ctx := r.Context()
	grpcReq := &pbslider.CreateSliderRequest{
		Name:  formData.Nama,
		Image: formData.FilePath,
	}

	res, err := h.client.Create(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "Create")
	}

	return httpx.JSON(w, http.StatusCreated, h.queryMapper.ToApiResponseSlider(res))
}

// @Security Bearer
// @Summary Update slider
// @Tags Slider Command
// @Description Update an existing slider's name or image
// @Accept mpfd
// @Produce json
// @Param id path int true "Slider ID"
// @Param name formData string false "Slider name"
// @Param image_slider formData file false "Slider image"
// @Success 200 {object} response.ApiResponseSlider "Successfully updated slider"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update slider"
// @Router /api/slider-command/update/{id} [post]
func (h *sliderCommandHandleApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	formData, err := h.parseSliderForm(w, r, false)
	if err != nil {
		return err
	}

	ctx := r.Context()
	grpcReq := &pbslider.UpdateSliderRequest{
		Id:    int32(id),
		Name:  formData.Nama,
		Image: formData.FilePath,
	}

	res, err := h.client.Update(ctx, grpcReq)
	if err != nil {
		return h.handleGrpcError(err, "Update")
	}

	h.cache.DeleteSliderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.queryMapper.ToApiResponseSlider(res))
}

// @Security Bearer
// @Summary Move slider to trash
// @Tags Slider Command
// @Description Move a slider record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Slider ID"
// @Success 200 {object} response.ApiResponseSliderDeleteAt "Successfully moved slider to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid slider ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move slider to trash"
// @Router /api/slider-command/trashed/{id} [post]
func (h *sliderCommandHandleApi) TrashedSlider(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedSlider(ctx, &pbslider.FindByIdSliderRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Trash")
	}

	h.cache.DeleteSliderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseSliderDeleteAt(res))
}

// @Security Bearer
// @Summary Restore trashed slider
// @Tags Slider Command
// @Description Restore a trashed slider record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Slider ID"
// @Success 200 {object} response.ApiResponseSliderDeleteAt "Successfully restored slider"
// @Failure 400 {object} errors.ErrorResponse "Invalid slider ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore slider"
// @Router /api/slider-command/restore/{id} [post]
func (h *sliderCommandHandleApi) RestoreSlider(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreSlider(ctx, &pbslider.FindByIdSliderRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Restore")
	}

	h.cache.DeleteSliderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseSliderDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete slider
// @Tags Slider Command
// @Description Permanently delete a slider record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Slider ID"
// @Success 200 {object} response.ApiResponseSliderDelete "Successfully deleted slider record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid slider ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete slider permanently"
// @Router /api/slider-command/permanent/{id} [delete]
func (h *sliderCommandHandleApi) DeleteSliderPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteSliderPermanent(ctx, &pbslider.FindByIdSliderRequest{Id: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Delete")
	}

	h.cache.DeleteSliderCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseSliderDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed sliders
// @Tags Slider Command
// @Description Restore all trashed slider records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseSliderAll "Successfully restored all sliders"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore sliders"
// @Router /api/slider-command/restore/all [post]
func (h *sliderCommandHandleApi) RestoreAllSlider(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllSlider(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "RestoreAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseSliderAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed sliders
// @Tags Slider Command
// @Description Permanently delete all trashed slider records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseSliderAll "Successfully deleted all sliders permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete sliders permanently"
// @Router /api/slider-command/permanent/all [post]
func (h *sliderCommandHandleApi) DeleteAllSliderPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllSliderPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "DeleteAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseSliderAll(res))
}

func (h *sliderCommandHandleApi) parseSliderForm(w http.ResponseWriter, r *http.Request, requireImage bool) (requests.SliderFormData, error) {
	var formData requests.SliderFormData

	formData.Nama = strings.TrimSpace(r.FormValue("name"))
	if formData.Nama == "" {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Name is required")
	}

	_, file, err := r.FormFile("image_slider")
	if err == nil {
		imagePath, err := h.upload.ProcessImageUpload(w, "uploads/slider", file, false)
		if err != nil {
			return formData, err
		}
		formData.FilePath = imagePath
	} else if requireImage {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Image is required")
	}

	return formData, nil
}

func (h *sliderCommandHandleApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to "+operation, zap.Error(err))
	return sharedErrors.ParseGrpcError(err)
}
