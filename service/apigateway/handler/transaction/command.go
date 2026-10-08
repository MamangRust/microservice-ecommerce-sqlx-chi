package transactionhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	transaction_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/transaction"
	pbtransaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type transactionCommandHandlerApi struct {
	client     pbtransaction.TransactionCommandServiceClient
	logger     logger.LoggerInterface
	mapper     apimapper.TransactionCommandResponseMapper
	cache      transaction_cache.TransactionCommandCache
	apiHandler apierror.ApiHandler
}

type transactionCommandHandleDeps struct {
	client     pbtransaction.TransactionCommandServiceClient
	router     chi.Router
	logger     logger.LoggerInterface
	mapper     apimapper.TransactionCommandResponseMapper
	cache      transaction_cache.TransactionCommandCache
	apiHandler apierror.ApiHandler
}

func NewTransactionCommandHandleApi(params *transactionCommandHandleDeps) *transactionCommandHandlerApi {
	handler := &transactionCommandHandlerApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	params.router.Route("/api/transaction-command", func(routerTransaction chi.Router) {
		routerTransaction.Post("/create", params.apiHandler.Handle("CreateTransaction", handler.Create))
		routerTransaction.Post("/update/{id}", params.apiHandler.Handle("UpdateTransaction", handler.Update))
		routerTransaction.Post("/trashed/{id}", params.apiHandler.Handle("TrashedTransaction", handler.Trashed))
		routerTransaction.Post("/restore/{id}", params.apiHandler.Handle("RestoreTransaction", handler.Restore))
		routerTransaction.Delete("/permanent/{id}", params.apiHandler.Handle("DeleteTransactionPermanent", handler.DeletePermanent))
		routerTransaction.Post("/restore/all", params.apiHandler.Handle("RestoreAllTransaction", handler.RestoreAll))
		routerTransaction.Post("/permanent/all", params.apiHandler.Handle("DeleteAllTransactionPermanent", handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new transaction
// @Tags Transaction Command
// @Description Create a new transaction with the provided details
// @Accept mpfd
// @Produce json
// @Param order_id formData int true "Order ID"
// @Param merchant_id formData int true "Merchant ID"
// @Param amount formData int true "Amount"
// @Param user_id formData int true "User ID"
// @Param status formData string true "Payment Status"
// @Param method formData string true "Payment Method"
// @Success 201 {object} response.ApiResponseTransaction "Successfully created transaction"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create transaction"
// @Router /api/transaction-command/create [post]
func (h *transactionCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	req := new(pbtransaction.CreateTransactionRequest)
	if err := httpx.Bind(r, req); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, req)
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.InvalidateTransactionCache(ctx)

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseTransaction(res))
}

// @Security Bearer
// @Summary Update transaction status
// @Tags Transaction Command
// @Description Update the payment status of a transaction
// @Accept mpfd
// @Produce json
// @Param id path int true "Transaction ID"
// @Param status formData string true "Payment Status"
// @Success 200 {object} response.ApiResponseTransaction "Successfully updated transaction"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update transaction"
// @Router /api/transaction-command/update/{id} [post]
func (h *transactionCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	status := r.FormValue("status")

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbtransaction.UpdateTransactionRequest{
		TransactionId: int32(id),
		PaymentStatus: status,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteTransactionCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransaction(res))
}

// @Security Bearer
// @Summary Move transaction to trash
// @Tags Transaction Command
// @Description Move a transaction record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.ApiResponseTransactionDeleteAt "Successfully moved transaction to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid transaction ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move transaction to trash"
// @Router /api/transaction-command/trashed/{id} [post]
func (h *transactionCommandHandlerApi) Trashed(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedTransaction(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteTransactionCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransactionDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed transaction
// @Tags Transaction Command
// @Description Restore a trashed transaction record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.ApiResponseTransactionDeleteAt "Successfully restored transaction"
// @Failure 400 {object} errors.ErrorResponse "Invalid transaction ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore transaction"
// @Router /api/transaction-command/restore/{id} [post]
func (h *transactionCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreTransaction(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteTransactionCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransactionDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a transaction
// @Tags Transaction Command
// @Description Permanently delete a transaction record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.ApiResponseTransactionDelete "Successfully deleted transaction record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid transaction ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete transaction permanently"
// @Router /api/transaction-command/permanent/{id} [delete]
func (h *transactionCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteTransactionPermanent(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(id)})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteTransactionCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransactionDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed transactions
// @Tags Transaction Command
// @Description Restore all trashed transaction records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTransactionAll "Successfully restored all transactions"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore transactions"
// @Router /api/transaction-command/restore/all [post]
func (h *transactionCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllTransaction(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.InvalidateTransactionCache(ctx)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransactionAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed transactions
// @Tags Transaction Command
// @Description Permanently delete all trashed transaction records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTransactionAll "Successfully deleted all transactions permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete transactions permanently"
// @Router /api/transaction-command/permanent/all [post]
func (h *transactionCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.InvalidateTransactionCache(ctx)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseTransactionAll(res))
}
