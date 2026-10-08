package merchanthandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchant_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantCommandHandlerApi struct {
	client       pbmerchant.MerchantCommandServiceClient
	logger       logger.LoggerInterface
	mapper       apimapper.MerchantCommandResponseMapper
	cache        merchant_cache.MerchantCommandCache
	upload_image upload_image.ImageUploads
	apiHandler   apierror.ApiHandler
}

type merchantCommandHandleDeps struct {
	client       pbmerchant.MerchantCommandServiceClient
	router       chi.Router
	logger       logger.LoggerInterface
	mapper       apimapper.MerchantCommandResponseMapper
	cache        merchant_cache.MerchantCommandCache
	upload_image upload_image.ImageUploads
	apiHandler   apierror.ApiHandler
}

func NewMerchantCommandHandleApi(params *merchantCommandHandleDeps) *merchantCommandHandlerApi {
	merchantCommandHandler := &merchantCommandHandlerApi{
		client:       params.client,
		logger:       params.logger,
		mapper:       params.mapper,
		cache:        params.cache,
		upload_image: params.upload_image,
		apiHandler:   params.apiHandler,
	}

	params.router.Route("/api/merchant-command", func(routerMerchant chi.Router) {

		routerMerchant.Post("/create", httpx.Handler(merchantCommandHandler.Create))
		routerMerchant.Post("/update/{id}", httpx.Handler(merchantCommandHandler.Update))
		routerMerchant.Post("/trashed/{id}", httpx.Handler(merchantCommandHandler.TrashedMerchant))
		routerMerchant.Post("/restore/{id}", httpx.Handler(merchantCommandHandler.RestoreMerchant))
		routerMerchant.Delete("/permanent/{id}", httpx.Handler(merchantCommandHandler.DeleteMerchantPermanent))
		routerMerchant.Post("/restore/all", httpx.Handler(merchantCommandHandler.RestoreAllMerchant))
		routerMerchant.Post("/permanent/all", httpx.Handler(merchantCommandHandler.DeleteAllMerchantPermanent))

	})
	return merchantCommandHandler
}

// @Security Bearer
// @Summary Create a new merchant
// @Tags Merchant Command
// @Description Create a new merchant for the current user
// @Accept json
// @Produce json
// @Param request body requests.CreateMerchantRequest true "Merchant details"
// @Success 201 {object} response.ApiResponseMerchant "Successfully created merchant"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant"
// @Router /api/merchant-command/create [post]
func (h *merchantCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return errors.ErrUnauthorized
	}

	var req requests.CreateMerchantRequest
	if err := httpx.Bind(r, &req); err != nil {
		return errors.NewBadRequestError("invalid request").WithInternal(err)
	}
	if err := req.Validate(); err != nil {
		return errors.NewValidationError(nil)
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchant.CreateMerchantRequest{
		UserId:       int32(userID),
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Status:       req.Status,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseMerchant(res))
}

// @Security Bearer
// @Summary Update an existing merchant
// @Tags Merchant Command
// @Description Update an existing merchant record
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Param request body requests.UpdateMerchantRequest true "Updated merchant details"
// @Success 200 {object} response.ApiResponseMerchant "Successfully updated merchant"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant"
// @Router /api/merchant-command/update/{id} [post]
func (h *merchantCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return errors.ErrUnauthorized
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	var req requests.UpdateMerchantRequest
	if err := httpx.Bind(r, &req); err != nil {
		return errors.NewBadRequestError("invalid request").WithInternal(err)
	}
	if err := req.Validate(); err != nil {
		return errors.NewValidationError(nil)
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchant.UpdateMerchantRequest{
		MerchantId:   int32(id),
		UserId:       int32(userID),
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		Status:       req.Status,
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedMerchant(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchant(res))
}

// @Security Bearer
// @Summary Move merchant to trash
// @Tags Merchant Command
// @Description Move a merchant record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} response.ApiResponseMerchantDeleteAt "Successfully moved merchant to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant to trash"
// @Router /api/merchant-command/trashed/{id} [post]
func (h *merchantCommandHandlerApi) TrashedMerchant(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedMerchant(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed merchant
// @Tags Merchant Command
// @Description Restore a trashed merchant record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} response.ApiResponseMerchant "Successfully restored merchant"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant"
// @Router /api/merchant-command/restore/{id} [post]
func (h *merchantCommandHandlerApi) RestoreMerchant(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.RestoreMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedMerchant(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchant(res))
}

// @Security Bearer
// @Summary Permanently delete a merchant
// @Tags Merchant Command
// @Description Permanently delete a merchant record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Merchant ID"
// @Success 200 {object} response.ApiResponseMerchantDelete "Successfully deleted merchant record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant permanently"
// @Router /api/merchant-command/permanent/{id} [delete]
func (h *merchantCommandHandlerApi) DeleteMerchantPermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.DeleteMerchantPermanent(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedMerchant(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed merchants
// @Tags Merchant Command
// @Description Restore all trashed merchant records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully restored all merchants"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchants"
// @Router /api/merchant-command/restore/all [post]
func (h *merchantCommandHandlerApi) RestoreAllMerchant(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllMerchant(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	// Invalidate all for bulk?
	// Normaly we'd clear everything related to merchants.

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchants
// @Tags Merchant Command
// @Description Permanently delete all trashed merchant records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully deleted all merchants permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchants permanently"
// @Router /api/merchant-command/permanent/all [post]
func (h *merchantCommandHandlerApi) DeleteAllMerchantPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAll(res))
}
