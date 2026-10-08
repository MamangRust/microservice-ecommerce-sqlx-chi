package merchantpolicyhandler

import (
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchantpolicy_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant_policies"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbmerchantpolicy "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type merchantPolicyQueryHandlerApi struct {
	client        pbmerchantpolicy.MerchantPolicyQueryServiceClient
	logger        logger.LoggerInterface
	mapper        apimapper.MerchantPolicyQueryResponseMapper
	cache         merchantpolicy_cache.MerchantPolicyQueryCache
	observability observability.TraceLoggerObservability
}

type merchantPolicyQueryHandleDeps struct {
	client        pbmerchantpolicy.MerchantPolicyQueryServiceClient
	router        chi.Router
	logger        logger.LoggerInterface
	mapper        apimapper.MerchantPolicyQueryResponseMapper
	cache         merchantpolicy_cache.MerchantPolicyQueryCache
	observability observability.TraceLoggerObservability
}

func NewMerchantPolicyQueryHandleApi(params *merchantPolicyQueryHandleDeps) *merchantPolicyQueryHandlerApi {
	handler := &merchantPolicyQueryHandlerApi{
		client:        params.client,
		logger:        params.logger,
		mapper:        params.mapper,
		cache:         params.cache,
		observability: params.observability,
	}

	params.router.Route("/api/merchant-policy-query", func(routerPolicy chi.Router) {
		routerPolicy.Get("/", httpx.Handler(handler.FindAll))
		routerPolicy.Get("/{id}", httpx.Handler(handler.FindById))
		routerPolicy.Get("/active", httpx.Handler(handler.FindByActive))
		routerPolicy.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all merchant policies
// @Tags Merchant Policy Query
// @Description Retrieve a list of all merchant policies
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantPolicies "List of merchant policies"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant policy data"
// @Router /api/merchant-policy-query [get]
func (h *merchantPolicyQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindAllMerchantPolicies")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindAll")
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantPolicies(res)
	h.cache.SetCachedMerchantPolicyAll(ctx, req, apiResponse)

	logSuccess("Successfully fetched all merchant policies")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find merchant policy by ID
// @Tags Merchant Policy Query
// @Description Retrieve a merchant policy by ID
// @Accept json
// @Produce json
// @Param id path int true "Policy ID"
// @Success 200 {object} response.ApiResponseMerchantPolicies "Merchant policy data"
// @Failure 400 {object} errors.ErrorResponse "Invalid policy ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant policy data"
// @Router /api/merchant-policy-query/{id} [get]
func (h *merchantPolicyQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByIdMerchantPolicy")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindById(ctx, &pbmerchantpolicy.FindByIdMerchantPoliciesRequest{Id: int32(id)})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindById")
	}

	apiResponse := h.mapper.ToApiResponseMerchantPolicies(res)
	h.cache.SetCachedMerchantPolicy(ctx, apiResponse)

	logSuccess("Successfully fetched merchant policy by ID")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active merchant policies
// @Tags Merchant Policy Query
// @Description Retrieve a list of active merchant policies
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantPoliciesDeleteAt "List of active merchant policies"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant policy data"
// @Router /api/merchant-policy-query/active [get]
func (h *merchantPolicyQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByActiveMerchantPolicies")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByActive")
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantPoliciesDeleteAt(res)
	h.cache.SetCachedMerchantPolicyActive(ctx, req, apiResponse)

	logSuccess("Successfully fetched active merchant policies")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed merchant policies
// @Tags Merchant Policy Query
// @Description Retrieve a list of trashed merchant policy records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantPoliciesDeleteAt "List of trashed merchant policy data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant policy data"
// @Router /api/merchant-policy-query/trashed [get]
func (h *merchantPolicyQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	req := &requests.FindAllMerchant{Page: page, PageSize: pageSize, Search: search}

	ctx, span, end, status, logSuccess := h.observability.StartTracingAndLogging(ctx, "FindByTrashedMerchantPolicies")
	defer func() {
		end(status)
	}()

	res, err := h.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		status = "error"
		return h.handleError(w, r, err, span, "FindByTrashed")
	}

	apiResponse := h.mapper.ToApiResponsePaginationMerchantPoliciesDeleteAt(res)
	h.cache.SetCachedMerchantPolicyTrashed(ctx, req, apiResponse)

	logSuccess("Successfully fetched trashed merchant policies")
	return httpx.JSON(w, http.StatusOK, apiResponse)
}

func (h *merchantPolicyQueryHandlerApi) handleError(w http.ResponseWriter, r *http.Request, err error, span trace.Span, method string) error {
	appErr := sharedErrors.ParseGrpcError(err)
	traceID := span.SpanContext().TraceID().String()

	h.logger.Error(
		"Merchant policy query error in "+method,
		zap.Error(err),
		zap.String("trace.id", traceID),
	)

	return apierror.HandleApiError(w, appErr, traceID)
}
