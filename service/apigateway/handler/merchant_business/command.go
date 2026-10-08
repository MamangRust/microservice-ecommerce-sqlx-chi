package merchantbusinesshandler

import (
	"net/http"
	"strconv"

	merchantbusiness_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_business"
	pbmerchantbusiness "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantBusinessCommandHandlerApi struct {
	client         pbmerchantbusiness.MerchantBusinessCommandServiceClient
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantBusinessCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantbusiness_cache.MerchantBusinessCommandCache
}

type merchantBusinessCommandHandleDeps struct {
	client         pbmerchantbusiness.MerchantBusinessCommandServiceClient
	router         chi.Router
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantBusinessCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantbusiness_cache.MerchantBusinessCommandCache
}

func NewMerchantBusinessCommandHandleApi(params *merchantBusinessCommandHandleDeps) *merchantBusinessCommandHandlerApi {
	handler := &merchantBusinessCommandHandlerApi{
		client:         params.client,
		logger:         params.logger,
		mapper:         params.mapper,
		merchantMapper: params.merchantMapper,
		cache:          params.cache,
	}

	params.router.Route("/api/merchant-business-command", func(routerBusiness chi.Router) {
		routerBusiness.Post("/create", httpx.Handler(handler.Create))
		routerBusiness.Post("/update/{id}", httpx.Handler(handler.Update))
		routerBusiness.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerBusiness.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerBusiness.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerBusiness.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerBusiness.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create merchant business information
// @Tags Merchant Business Command
// @Description Create business information for a merchant
// @Accept json
// @Produce json
// @Param body body requests.CreateMerchantBusinessInformationRequest true "Create merchant business request"
// @Success 200 {object} response.ApiResponseMerchantBusiness "Successfully created merchant business info"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant business info"
// @Router /api/merchant-business-command/create [post]
func (h *merchantBusinessCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateMerchantBusinessInformationRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchantbusiness.CreateMerchantBusinessRequest{
		MerchantId:        int32(body.MerchantID),
		BusinessType:      body.BusinessType,
		TaxId:             body.TaxID,
		EstablishedYear:   int32(body.EstablishedYear),
		NumberOfEmployees: int32(body.NumberOfEmployees),
		WebsiteUrl:        body.WebsiteUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantBusiness(res))
}

// @Security Bearer
// @Summary Update merchant business information
// @Tags Merchant Business Command
// @Description Update existing business information for a merchant
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Param body body requests.UpdateMerchantBusinessInformationRequest true "Update merchant business request"
// @Success 200 {object} response.ApiResponseMerchantBusiness "Successfully updated merchant business info"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant business info"
// @Router /api/merchant-business-command/update/{id} [post]
func (h *merchantBusinessCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateMerchantBusinessInformationRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	body.MerchantBusinessInfoID = &id
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchantbusiness.UpdateMerchantBusinessRequest{
		MerchantBusinessInfoId: int32(id),
		BusinessType:           body.BusinessType,
		TaxId:                  body.TaxID,
		EstablishedYear:        int32(body.EstablishedYear),
		NumberOfEmployees:      int32(body.NumberOfEmployees),
		WebsiteUrl:             body.WebsiteUrl,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantBusinessCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantBusiness(res))
}

// @Security Bearer
// @Summary Move merchant business info to trash
// @Tags Merchant Business Command
// @Description Move a merchant business info record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 200 {object} response.ApiResponseMerchantBusinessDeleteAt "Successfully moved merchant business info to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid business ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant business info to trash"
// @Router /api/merchant-business-command/trashed/{id} [post]
func (h *merchantBusinessCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedMerchantBusiness(ctx, &pbmerchantbusiness.FindByIdMerchantBusinessRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantBusinessCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantBusinessDeleteAt(res))
}

// @Security Bearer
// @Summary Restore trashed merchant business info
// @Tags Merchant Business Command
// @Description Restore a trashed merchant business info record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 200 {object} response.ApiResponseMerchantBusinessDeleteAt "Successfully restored merchant business info"
// @Failure 400 {object} errors.ErrorResponse "Invalid business ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant business info"
// @Router /api/merchant-business-command/restore/{id} [post]
func (h *merchantBusinessCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreMerchantBusiness(ctx, &pbmerchantbusiness.FindByIdMerchantBusinessRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantBusinessCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantBusinessDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete merchant business info
// @Tags Merchant Business Command
// @Description Permanently delete a merchant business info record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Business ID"
// @Success 200 {object} response.ApiResponseMerchantDelete "Successfully deleted merchant business info record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid business ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant business info permanently"
// @Router /api/merchant-business-command/permanent/{id} [delete]
func (h *merchantBusinessCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteMerchantBusinessPermanent(ctx, &pbmerchantbusiness.FindByIdMerchantBusinessRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteMerchantBusinessCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed merchant business info
// @Tags Merchant Business Command
// @Description Restore all trashed merchant business info records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully restored all merchant business info"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant business info"
// @Router /api/merchant-business-command/restore/all [post]
func (h *merchantBusinessCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllMerchantBusiness(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchant business info
// @Tags Merchant Business Command
// @Description Permanently delete all trashed merchant business info records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully deleted all merchant business info permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant business info permanently"
// @Router /api/merchant-business-command/permanent/all [post]
func (h *merchantBusinessCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllMerchantBusinessPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}
