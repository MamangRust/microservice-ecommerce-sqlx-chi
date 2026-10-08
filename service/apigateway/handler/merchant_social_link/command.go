package merchantsociallinkhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_social_link"
	pbmerchantsociallink "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_social_link"
	"github.com/go-chi/chi/v5"
)

type merchantSocialLinkCommandHandlerApi struct {
	client     pbmerchantsociallink.MerchantSocialCommandServiceClient
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantSocialLinkCommandResponseMapper
	apiHandler apierror.ApiHandler
}

type merchantSocialLinkCommandHandleDeps struct {
	client     pbmerchantsociallink.MerchantSocialCommandServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.MerchantSocialLinkCommandResponseMapper
	apiHandler apierror.ApiHandler
}

func NewMerchantSocialLinkCommandHandleApi(params *merchantSocialLinkCommandHandleDeps) *merchantSocialLinkCommandHandlerApi {
	handler := &merchantSocialLinkCommandHandlerApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/merchant-social-link", func(routerSocial chi.Router) {
		routerSocial.Post("/create", httpx.Handler(handler.Create))
		routerSocial.Post("/update/{id}", httpx.Handler(handler.Update))

	})
	return handler
}

// @Security Bearer
// @Summary Create merchant social link
// @Tags Merchant Social Link Command
// @Description Create a new social media link for a merchant
// @Accept json
// @Produce json
// @Param body body requests.CreateMerchantSocialRequest true "Create merchant social link request"
// @Success 200 {object} response.ApiResponseMerchantSocialLink "Successfully created merchant social link"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant social link"
// @Router /api/merchant-social-link/create [post]
func (h *merchantSocialLinkCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateMerchantSocialRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchantsociallink.CreateMerchantSocialRequest{
		MerchantDetailId: int32(*body.MerchantDetailID),
		Platform:         body.Platform,
		Url:              body.Url,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantSocialLink(res))
}

// @Security Bearer
// @Summary Update merchant social link
// @Tags Merchant Social Link Command
// @Description Update an existing social media link for a merchant
// @Accept json
// @Produce json
// @Param id path int true "Link ID"
// @Param body body requests.UpdateMerchantSocialRequest true "Update merchant social link request"
// @Success 200 {object} response.ApiResponseMerchantSocialLink "Successfully updated merchant social link"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant social link"
// @Router /api/merchant-social-link/update/{id} [post]
func (h *merchantSocialLinkCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateMerchantSocialRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchantsociallink.UpdateMerchantSocialRequest{
		Id:               int32(id),
		MerchantDetailId: int32(*body.MerchantDetailID),
		Platform:         body.Platform,
		Url:              body.Url,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantSocialLink(res))
}
