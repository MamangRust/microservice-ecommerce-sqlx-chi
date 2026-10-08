package merchantdetailhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchant_detail_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_detail"
	pbmerchantdetail "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantDetailCommandHandlerApi struct {
	client         pbmerchantdetail.MerchantDetailCommandServiceClient
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantDetailCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchant_detail_cache.MerchantDetailCommandCache
	upload_image   upload_image.ImageUploads
	apiHandler     apierror.ApiHandler
}

type merchantDetailCommandHandleDeps struct {
	client         pbmerchantdetail.MerchantDetailCommandServiceClient
	router         chi.Router
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantDetailCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchant_detail_cache.MerchantDetailCommandCache
	upload_image   upload_image.ImageUploads
	apiHandler     apierror.ApiHandler
}

func NewMerchantDetailCommandHandleApi(params *merchantDetailCommandHandleDeps) *merchantDetailCommandHandlerApi {
	handler := &merchantDetailCommandHandlerApi{
		client:         params.client,
		logger:         params.logger,
		mapper:         params.mapper,
		merchantMapper: params.merchantMapper,
		cache:          params.cache,
		upload_image:   params.upload_image,
		apiHandler:     params.apiHandler,
	}

	params.router.Route("/api/merchant-detail-command", func(router chi.Router) {
		router.Post("/create", httpx.Handler(handler.Create))
		router.Post("/update/{id}", httpx.Handler(handler.Update))
		router.Post("/trashed/{id}", httpx.Handler(handler.Trashed))
		router.Post("/restore/{id}", httpx.Handler(handler.Restore))
		router.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		router.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		router.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create merchant details
// @Tags Merchant Detail Command
// @Description Create detailed information for a merchant
// @Accept json
// @Produce json
// @Param body body requests.CreateMerchantDetailRequest true "Create merchant detail request"
// @Success 201 {object} response.ApiResponseMerchantDetail "Successfully created merchant detail"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant detail"
// @Router /api/merchant-detail-command/create [post]
func (h *merchantDetailCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var req requests.CreateMerchantDetailRequest
	if err := httpx.Bind(r, &req); err != nil {
		return sharedErrors.NewBadRequestError("invalid request").WithInternal(err)
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchantdetail.CreateMerchantDetailRequest{
		MerchantId:       int32(req.MerchantID),
		DisplayName:      req.DisplayName,
		CoverImageUrl:    req.CoverImageUrl,
		LogoUrl:          req.LogoUrl,
		ShortDescription: req.ShortDescription,
		WebsiteUrl:       req.WebsiteUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseMerchantDetail(res))
}

// @Security Bearer
// @Summary Update merchant details
// @Tags Merchant Detail Command
// @Description Update existing detailed information for a merchant
// @Accept json
// @Produce json
// @Param id path int true "Detail ID"
// @Param body body requests.UpdateMerchantDetailRequest true "Update merchant detail request"
// @Success 200 {object} response.ApiResponseMerchantDetail "Successfully updated merchant detail"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant detail"
// @Router /api/merchant-detail-command/update/{id} [post]
func (h *merchantDetailCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return sharedErrors.NewBadRequestError("id is required")
	}

	var req requests.UpdateMerchantDetailRequest
	if err := httpx.Bind(r, &req); err != nil {
		return sharedErrors.NewBadRequestError("invalid request").WithInternal(err)
	}
	req.MerchantDetailID = &id
	if err := req.Validate(); err != nil {
		return sharedErrors.NewBadRequestError(err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchantdetail.UpdateMerchantDetailRequest{
		MerchantDetailId: int32(id),
		DisplayName:      req.DisplayName,
		CoverImageUrl:    req.CoverImageUrl,
		LogoUrl:          req.LogoUrl,
		ShortDescription: req.ShortDescription,
		WebsiteUrl:       req.WebsiteUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantDetailCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDetail(res))
}

// @Security Bearer
// @Summary Move merchant detail to trash
// @Tags Merchant Detail Command
// @Description Move a merchant detail record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Detail ID"
// @Success 200 {object} response.ApiResponseMerchantDetailDeleteAt "Successfully moved merchant detail to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant detail to trash"
// @Router /api/merchant-detail-command/trashed/{id} [post]
func (h *merchantDetailCommandHandlerApi) Trashed(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return sharedErrors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.TrashedMerchantDetail(ctx, &pbmerchantdetail.FindByIdMerchantDetailRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantDetailCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDetailDeleteAt(res))
}

// @Security Bearer
// @Summary Restore trashed merchant detail
// @Tags Merchant Detail Command
// @Description Restore a trashed merchant detail record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Detail ID"
// @Success 200 {object} response.ApiResponseMerchantDetailDeleteAt "Successfully restored merchant detail"
// @Failure 400 {object} errors.ErrorResponse "Invalid detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant detail"
// @Router /api/merchant-detail-command/restore/{id} [post]
func (h *merchantDetailCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return sharedErrors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.RestoreMerchantDetail(ctx, &pbmerchantdetail.FindByIdMerchantDetailRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantDetailCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDetailDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete merchant detail
// @Tags Merchant Detail Command
// @Description Permanently delete a merchant detail record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Detail ID"
// @Success 204 "Successfully deleted merchant detail record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid detail ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant detail permanently"
// @Router /api/merchant-detail-command/permanent/{id} [delete]
func (h *merchantDetailCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return sharedErrors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.DeleteMerchantDetailPermanent(ctx, &pbmerchantdetail.FindByIdMerchantDetailRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantDetailCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed merchant details
// @Tags Merchant Detail Command
// @Description Restore all trashed merchant detail records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully restored all merchant details"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant details"
// @Router /api/merchant-detail-command/restore/all [post]
func (h *merchantDetailCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllMerchantDetail(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchant details
// @Tags Merchant Detail Command
// @Description Permanently delete all trashed merchant detail records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully deleted all merchant details permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant details permanently"
// @Router /api/merchant-detail-command/permanent/all [post]
func (h *merchantDetailCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllMerchantDetailPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}
