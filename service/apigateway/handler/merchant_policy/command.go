package merchantpolicyhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchantpolicy_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_policies"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchantapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbmerchantpolicy "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantPolicyCommandHandlerApi struct {
	client         pbmerchantpolicy.MerchantPolicyCommandServiceClient
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantPolicyCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantpolicy_cache.MerchantPolicyCommandCache
	observability  observability.TraceLoggerObservability
}

type merchantPolicyCommandHandleDeps struct {
	client         pbmerchantpolicy.MerchantPolicyCommandServiceClient
	router         chi.Router
	logger         logger.LoggerInterface
	mapper         apimapper.MerchantPolicyCommandResponseMapper
	merchantMapper merchantapimapper.MerchantCommandResponseMapper
	cache          merchantpolicy_cache.MerchantPolicyCommandCache
	observability  observability.TraceLoggerObservability
}

func NewMerchantPolicyCommandHandleApi(params *merchantPolicyCommandHandleDeps) *merchantPolicyCommandHandlerApi {
	handler := &merchantPolicyCommandHandlerApi{
		client:         params.client,
		logger:         params.logger,
		mapper:         params.mapper,
		merchantMapper: params.merchantMapper,
		cache:          params.cache,
		observability:  params.observability,
	}

	params.router.Route("/api/merchant-policy-command", func(routerPolicy chi.Router) {
		routerPolicy.Post("/create", httpx.Handler(handler.Create))
		routerPolicy.Post("/update/{id}", httpx.Handler(handler.Update))
		routerPolicy.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerPolicy.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerPolicy.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerPolicy.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerPolicy.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create merchant policy
// @Tags Merchant Policy Command
// @Description Create a new policy for a merchant
// @Accept json
// @Produce json
// @Param body body requests.CreateMerchantPolicyRequest true "Create merchant policy request"
// @Success 200 {object} response.ApiResponseMerchantPolicies "Successfully created merchant policy"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant policy"
// @Router /api/merchant-policy-command/create [post]
func (h *merchantPolicyCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var body requests.CreateMerchantPolicyRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "CreateMerchantPolicy")
	defer func() {
		end(status)
	}()

	res, err := h.client.Create(ctx, &pbmerchantpolicy.CreateMerchantPoliciesRequest{
		MerchantId:  int32(body.MerchantID),
		PolicyType:  body.PolicyType,
		Title:       body.Title,
		Description: body.Description,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Create")
	}

	logSuccess("Successfully created merchant policy")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantPolicies(res))
}

// @Security Bearer
// @Summary Update merchant policy
// @Tags Merchant Policy Command
// @Description Update an existing policy for a merchant
// @Accept json
// @Produce json
// @Param id path int true "Policy ID"
// @Param body body requests.UpdateMerchantPolicyRequest true "Update merchant policy request"
// @Success 200 {object} response.ApiResponseMerchantPolicies "Successfully updated merchant policy"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant policy"
// @Router /api/merchant-policy-command/update/{id} [post]
func (h *merchantPolicyCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateMerchantPolicyRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	body.MerchantPolicyID = &id
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "UpdateMerchantPolicy")
	defer func() {
		end(status)
	}()

	res, err := h.client.Update(ctx, &pbmerchantpolicy.UpdateMerchantPoliciesRequest{
		MerchantPolicyId: int32(id),
		PolicyType:       body.PolicyType,
		Title:            body.Title,
		Description:      body.Description,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Update")
	}

	h.cache.DeleteMerchantPolicyCache(ctx, 0) // Body has no MerchantID

	logSuccess("Successfully updated merchant policy")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantPolicies(res))
}

// @Security Bearer
// @Summary Move merchant policy to trash
// @Tags Merchant Policy Command
// @Description Move a merchant policy record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Policy ID"
// @Success 200 {object} response.ApiResponseMerchantPoliciesDeleteAt "Successfully moved merchant policy to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid policy ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant policy to trash"
// @Router /api/merchant-policy-command/trashed/{id} [post]
func (h *merchantPolicyCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "TrashedMerchantPolicy")
	defer func() {
		end(status)
	}()

	res, err := h.client.TrashedMerchantPolicies(ctx, &pbmerchantpolicy.FindByIdMerchantPoliciesRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Trash")
	}

	h.cache.DeleteMerchantPolicyCache(ctx, 0)

	logSuccess("Successfully moved merchant policy to trash")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantPoliciesDeleteAt(res))
}

// @Security Bearer
// @Summary Restore trashed merchant policy
// @Tags Merchant Policy Command
// @Description Restore a trashed merchant policy record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Policy ID"
// @Success 200 {object} response.ApiResponseMerchantPoliciesDeleteAt "Successfully restored merchant policy"
// @Failure 400 {object} errors.ErrorResponse "Invalid policy ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant policy"
// @Router /api/merchant-policy-command/restore/{id} [post]
func (h *merchantPolicyCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreMerchantPolicy")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreMerchantPolicies(ctx, &pbmerchantpolicy.FindByIdMerchantPoliciesRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "Restore")
	}

	h.cache.DeleteMerchantPolicyCache(ctx, 0)

	logSuccess("Successfully restored merchant policy")
	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantPoliciesDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete merchant policy
// @Tags Merchant Policy Command
// @Description Permanently delete a merchant policy record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Policy ID"
// @Success 200 {object} response.ApiResponseMerchantDelete "Successfully deleted merchant policy record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid policy ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant policy permanently"
// @Router /api/merchant-policy-command/permanent/{id} [delete]
func (h *merchantPolicyCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteMerchantPolicyPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteMerchantPoliciesPermanent(ctx, &pbmerchantpolicy.FindByIdMerchantPoliciesRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "DeletePermanent")
	}

	h.cache.DeleteMerchantPolicyCache(ctx, 0)

	logSuccess("Successfully deleted merchant policy permanently")
	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed merchant policies
// @Tags Merchant Policy Command
// @Description Restore all trashed merchant policy records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully restored all merchant policies"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant policies"
// @Router /api/merchant-policy-command/restore/all [post]
func (h *merchantPolicyCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "RestoreAllMerchantPolicies")
	defer func() {
		end(status)
	}()

	res, err := h.client.RestoreAllMerchantPolicies(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "RestoreAll")
	}

	logSuccess("Successfully restored all merchant policies")
	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchant policies
// @Tags Merchant Policy Command
// @Description Permanently delete all trashed merchant policy records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantAll "Successfully deleted all merchant policies permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant policies permanently"
// @Router /api/merchant-policy-command/permanent/all [post]
func (h *merchantPolicyCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "DeleteAllMerchantPoliciesPermanent")
	defer func() {
		end(status)
	}()

	res, err := h.client.DeleteAllMerchantPoliciesPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "DeleteAllPermanent")
	}

	logSuccess("Successfully deleted all merchant policies permanently")
	return httpx.JSON(w, http.StatusOK, h.merchantMapper.ToApiResponseMerchantAll(res))
}

func (h *merchantPolicyCommandHandlerApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		"Merchant policy command error in "+method,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
