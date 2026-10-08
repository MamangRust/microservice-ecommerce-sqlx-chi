package transactionhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	transaction_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/transaction"
	pbtransaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/go-chi/chi/v5"
)

type transactionQueryHandlerApi struct {
	queryClient pbtransaction.TransactionQueryServiceClient
	logger      logger.LoggerInterface
	mapper      apimapper.TransactionQueryResponseMapper
	cache       transaction_cache.TransactionQueryCache
	apiHandler  apierror.ApiHandler
}

type transactionQueryHandleDeps struct {
	queryClient pbtransaction.TransactionQueryServiceClient
	router      chi.Router
	logger      logger.LoggerInterface
	mapper      apimapper.TransactionQueryResponseMapper
	cache       transaction_cache.TransactionQueryCache
	apiHandler  apierror.ApiHandler
}

func NewTransactionQueryHandleApi(params *transactionQueryHandleDeps) *transactionQueryHandlerApi {
	handler := &transactionQueryHandlerApi{
		queryClient: params.queryClient,
		logger:      params.logger,
		mapper:      params.mapper,
		cache:       params.cache,
		apiHandler:  params.apiHandler,
	}

	params.router.Route("/api/transaction-query", func(routerTransaction chi.Router) {
		routerTransaction.Get("/", params.apiHandler.Handle("FindAllTransactions", handler.FindAll))
		routerTransaction.Get("/{id}", params.apiHandler.Handle("FindById", handler.FindById))
		routerTransaction.Get("/merchant/{merchant_id}", params.apiHandler.Handle("FindByMerchant", handler.FindByMerchant))
		routerTransaction.Get("/active", params.apiHandler.Handle("FindByActive", handler.FindByActive))
		routerTransaction.Get("/trashed", params.apiHandler.Handle("FindByTrashed", handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all transactions
// @Tags Transaction Query
// @Description Retrieve a list of all transactions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationTransaction "List of transactions"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve transaction data"
// @Router /api/transaction-query [get]
func (h *transactionQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllTransaction{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedTransactionsCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.queryClient.FindAllTransactions(ctx, &pbtransaction.FindAllTransactionRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationTransaction(res)
	h.cache.SetCachedTransactionsCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find transaction by ID
// @Tags Transaction Query
// @Description Retrieve a transaction by ID
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.ApiResponseTransaction "Transaction data"
// @Failure 400 {object} errors.ErrorResponse "Invalid transaction ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve transaction data"
// @Router /api/transaction-query/{id} [get]
func (h *transactionQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Transaction ID")
	}

	ctx := r.Context()
	if cachedData, found := h.cache.GetCachedTransactionCache(ctx, id); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.queryClient.FindById(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseTransaction(res)
	h.cache.SetCachedTransactionCache(ctx, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Find transactions by merchant ID
// @Tags Transaction Query
// @Description Retrieve a list of transactions belonging to a specific merchant
// @Accept json
// @Produce json
// @Param merchant_id path int true "Merchant ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationTransaction "List of transactions by merchant"
// @Failure 400 {object} errors.ErrorResponse "Invalid merchant ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve transaction data"
// @Router /api/transaction-query/merchant/{merchant_id} [get]
func (h *transactionQueryHandlerApi) FindByMerchant(w http.ResponseWriter, r *http.Request) error {
	merchantID, err := strconv.Atoi(chi.URLParam(r, "merchant_id"))
	if err != nil || merchantID <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid Merchant ID")
	}

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
	req := &requests.FindAllTransactionByMerchant{MerchantID: merchantID, Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedTransactionByMerchant(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.queryClient.FindByMerchant(ctx, &pbtransaction.FindAllTransactionByMerchantRequest{
		MerchantId: int32(merchantID), Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationTransaction(res)
	h.cache.SetCachedTransactionByMerchant(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve active transactions
// @Tags Transaction Query
// @Description Retrieve a list of active transactions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationTransaction "List of active transactions"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve transaction data"
// @Router /api/transaction-query/active [get]
func (h *transactionQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllTransaction{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedTransactionActiveCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.queryClient.FindByActive(ctx, &pbtransaction.FindAllTransactionRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationTransaction(res)
	h.cache.SetCachedTransactionActiveCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}

// @Security Bearer
// @Summary Retrieve trashed transactions
// @Tags Transaction Query
// @Description Retrieve a list of trashed transaction records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationTransactionDeleteAt "List of trashed transaction data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve transaction data"
// @Router /api/transaction-query/trashed [get]
func (h *transactionQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
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
	req := &requests.FindAllTransaction{Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedTransactionTrashedCache(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.queryClient.FindByTrashed(ctx, &pbtransaction.FindAllTransactionRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponsePaginationTransactionDeleteAt(res)
	h.cache.SetCachedTransactionTrashedCache(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
