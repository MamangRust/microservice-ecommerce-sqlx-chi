package merchantawardhandler

import (
	"net/http"
	"strconv"

	merchantawards_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_awards"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_award"
	pbmerchantaward "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_award"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantAwardCommandHandlerApi struct {
	client         pbmerchantaward.MerchantAwardCommandServiceClient
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantAwardCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantawards_cache.MerchantAwardCommandCache
}

type merchantAwardCommandHandleDeps struct {
	client         pbmerchantaward.MerchantAwardCommandServiceClient
	router         chi.Router
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantAwardCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantawards_cache.MerchantAwardCommandCache
}

func NewMerchantAwardCommandHandleApi(params *merchantAwardCommandHandleDeps) *merchantAwardCommandHandlerApi {
	handler := &merchantAwardCommandHandlerApi{
		client:         params.client,
		logger:         params.logger,
		mapper:         params.mapper,
		merchantMapper: params.merchantMapper,
		cache:          params.cache,
	}

	params.router.Route("/api/merchant-award-command", func(routerAward chi.Router) {
		routerAward.Post("/create", httpx.Handler(handler.Create))
		routerAward.Post("/update/{id}", httpx.Handler(handler.Update))
		routerAward.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerAward.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerAward.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerAward.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerAward.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new merchant award
// @Tags Merchant Award Command
// @Description Create a new merchant award or certification record
// @Accept json
// @Produce json
// @Param body body requests.CreateMerchantCertificationOrAwardRequest true "Create merchant award request"
// @Success 200 {object} response.ApiResponseMerchantAward "Successfully created merchant award"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant award"
// @Router /api/merchant-award-command/create [post]
func (h *merchantAwardCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateMerchantCertificationOrAwardRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchantaward.CreateMerchantAwardRequest{
		MerchantId:     int32(body.MerchantID),
		Title:          body.Title,
		Description:    body.Description,
		IssuedBy:       body.IssuedBy,
		IssueDate:      body.IssueDate,
		ExpiryDate:     body.ExpiryDate,
		CertificateUrl: body.CertificateUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAward(res))
}

// @Security Bearer
// @Summary Update merchant award
// @Tags Merchant Award Command
// @Description Update an existing merchant award or certification record
// @Accept json
// @Produce json
// @Param id path int true "Award ID"
// @Param body body requests.UpdateMerchantCertificationOrAwardRequest true "Update merchant award request"
// @Success 200 {object} response.ApiResponseMerchantAward "Successfully updated merchant award"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant award"
// @Router /api/merchant-award-command/update/{id} [post]
func (h *merchantAwardCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateMerchantCertificationOrAwardRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	body.MerchantCertificationID = &id
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchantaward.UpdateMerchantAwardRequest{
		MerchantCertificationId: int32(id),
		Title:                   body.Title,
		Description:             body.Description,
		IssuedBy:                body.IssuedBy,
		IssueDate:               body.IssueDate,
		ExpiryDate:              body.ExpiryDate,
		CertificateUrl:          body.CertificateUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantAwardCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAward(res))
}

// @Security Bearer
// @Summary Move merchant award to trash
// @Tags Merchant Award Command
// @Description Move a merchant award record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Award ID"
// @Success 200 {object} response.ApiResponseMerchantAwardDeleteAt "Successfully moved merchant award to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid award ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant award to trash"
// @Router /api/merchant-award-command/trashed/{id} [post]
func (h *merchantAwardCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedMerchantAward(ctx, &pbmerchantaward.FindByIdMerchantAwardRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantAwardCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAwardDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed merchant award
// @Tags Merchant Award Command
// @Description Restore a trashed merchant award record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Award ID"
// @Success 200 {object} response.ApiResponseMerchantAwardDeleteAt "Successfully restored merchant award"
// @Failure 400 {object} errors.ErrorResponse "Invalid award ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant award"
// @Router /api/merchant-award-command/restore/{id} [post]
func (h *merchantAwardCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreMerchantAward(ctx, &pbmerchantaward.FindByIdMerchantAwardRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantAwardCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantAwardDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a merchant award
// @Tags Merchant Award Command
// @Description Permanently delete a merchant award record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Award ID"
// @Success 200 {object} response.ApiResponseMerchantDelete "Successfully deleted merchant award record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid award ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant award permanently"
// @Router /api/merchant-award-command/permanent/{id} [delete]
func (h *merchantAwardCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteMerchantAwardPermanent(ctx, &pbmerchantaward.FindByIdMerchantAwardRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantAwardCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed merchant awards
// @Tags Merchant Award Command
// @Description Restore all trashed merchant award records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully restored all merchant awards"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant awards"
// @Router /api/merchant-award-command/restore/all [post]
func (h *merchantAwardCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllMerchantAward(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchant awards
// @Tags Merchant Award Command
// @Description Permanently delete all trashed merchant award records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully deleted all merchant awards permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant awards permanently"
// @Router /api/merchant-award-command/permanent/all [post]
func (h *merchantAwardCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllMerchantAwardPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}
