package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	transactionrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/transaction"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"go.uber.org/zap"
)

// TransactionStatsHandler serves TransactionStatsService and
// TransactionStatsByMerchantService from ClickHouse.
type TransactionStatsHandler struct {
	pb_transaction.UnimplementedTransactionStatsServiceServer
	pb_transaction.UnimplementedTransactionStatsByMerchantServiceServer
	repo transactionrepo.Repository
	log  logger.LoggerInterface
}

func NewTransactionStatsHandler(repo transactionrepo.Repository, log logger.LoggerInterface) *TransactionStatsHandler {
	return &TransactionStatsHandler{repo: repo, log: log}
}

// --- TransactionStatsService ---

func (h *TransactionStatsHandler) GetMonthlyAmountSuccess(ctx context.Context, req *pb_transaction.MonthAmountTransactionRequest) (*pb_transaction.ApiResponseTransactionMonthAmountSuccess, error) {
	data, err := h.repo.GetMonthlyAmount(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusSuccess, 0)
	if err != nil {
		h.log.Error("GetMonthlyAmountSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthAmountSuccess{
		Status:  "success",
		Message: "Successfully fetched monthly amount success stats",
		Data:    mapMonthlyAmountSuccess(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyAmountSuccess(ctx context.Context, req *pb_transaction.YearAmountTransactionRequest) (*pb_transaction.ApiResponseTransactionYearAmountSuccess, error) {
	data, err := h.repo.GetYearlyAmount(ctx, int(req.GetYear()), transactionrepo.StatusSuccess, 0)
	if err != nil {
		h.log.Error("GetYearlyAmountSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearAmountSuccess{
		Status:  "success",
		Message: "Successfully fetched yearly amount success stats",
		Data:    mapYearlyAmountSuccess(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyAmountFailed(ctx context.Context, req *pb_transaction.MonthAmountTransactionRequest) (*pb_transaction.ApiResponseTransactionMonthAmountFailed, error) {
	data, err := h.repo.GetMonthlyAmount(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusFailed, 0)
	if err != nil {
		h.log.Error("GetMonthlyAmountFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthAmountFailed{
		Status:  "success",
		Message: "Successfully fetched monthly amount failed stats",
		Data:    mapMonthlyAmountFailed(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyAmountFailed(ctx context.Context, req *pb_transaction.YearAmountTransactionRequest) (*pb_transaction.ApiResponseTransactionYearAmountFailed, error) {
	data, err := h.repo.GetYearlyAmount(ctx, int(req.GetYear()), transactionrepo.StatusFailed, 0)
	if err != nil {
		h.log.Error("GetYearlyAmountFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearAmountFailed{
		Status:  "success",
		Message: "Successfully fetched yearly amount failed stats",
		Data:    mapYearlyAmountFailed(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyTransactionMethodSuccess(ctx context.Context, req *pb_transaction.MonthMethodTransactionRequest) (*pb_transaction.ApiResponseTransactionMonthPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethod(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusSuccess, 0)
	if err != nil {
		h.log.Error("GetMonthlyTransactionMethodSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Successfully fetched monthly transaction method success stats",
		Data:    mapMonthlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyTransactionMethodSuccess(ctx context.Context, req *pb_transaction.YearMethodTransactionRequest) (*pb_transaction.ApiResponseTransactionYearPaymentmethod, error) {
	data, err := h.repo.GetYearlyMethod(ctx, int(req.GetYear()), transactionrepo.StatusSuccess, 0)
	if err != nil {
		h.log.Error("GetYearlyTransactionMethodSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Successfully fetched yearly transaction method success stats",
		Data:    mapYearlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyTransactionMethodFailed(ctx context.Context, req *pb_transaction.MonthMethodTransactionRequest) (*pb_transaction.ApiResponseTransactionMonthPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethod(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusFailed, 0)
	if err != nil {
		h.log.Error("GetMonthlyTransactionMethodFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Successfully fetched monthly transaction method failed stats",
		Data:    mapMonthlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyTransactionMethodFailed(ctx context.Context, req *pb_transaction.YearMethodTransactionRequest) (*pb_transaction.ApiResponseTransactionYearPaymentmethod, error) {
	data, err := h.repo.GetYearlyMethod(ctx, int(req.GetYear()), transactionrepo.StatusFailed, 0)
	if err != nil {
		h.log.Error("GetYearlyTransactionMethodFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Successfully fetched yearly transaction method failed stats",
		Data:    mapYearlyMethod(data),
	}, nil
}

// --- TransactionStatsByMerchantService ---

func (h *TransactionStatsHandler) GetMonthlyAmountSuccessByMerchant(ctx context.Context, req *pb_transaction.MonthAmountTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionMonthAmountSuccess, error) {
	data, err := h.repo.GetMonthlyAmount(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusSuccess, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetMonthlyAmountSuccessByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthAmountSuccess{
		Status:  "success",
		Message: "Successfully fetched monthly amount success stats by merchant",
		Data:    mapMonthlyAmountSuccess(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyAmountSuccessByMerchant(ctx context.Context, req *pb_transaction.YearAmountTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionYearAmountSuccess, error) {
	data, err := h.repo.GetYearlyAmount(ctx, int(req.GetYear()), transactionrepo.StatusSuccess, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetYearlyAmountSuccessByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearAmountSuccess{
		Status:  "success",
		Message: "Successfully fetched yearly amount success stats by merchant",
		Data:    mapYearlyAmountSuccess(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyAmountFailedByMerchant(ctx context.Context, req *pb_transaction.MonthAmountTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionMonthAmountFailed, error) {
	data, err := h.repo.GetMonthlyAmount(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusFailed, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetMonthlyAmountFailedByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthAmountFailed{
		Status:  "success",
		Message: "Successfully fetched monthly amount failed stats by merchant",
		Data:    mapMonthlyAmountFailed(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyAmountFailedByMerchant(ctx context.Context, req *pb_transaction.YearAmountTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionYearAmountFailed, error) {
	data, err := h.repo.GetYearlyAmount(ctx, int(req.GetYear()), transactionrepo.StatusFailed, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetYearlyAmountFailedByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearAmountFailed{
		Status:  "success",
		Message: "Successfully fetched yearly amount failed stats by merchant",
		Data:    mapYearlyAmountFailed(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyTransactionMethodByMerchantSuccess(ctx context.Context, req *pb_transaction.MonthMethodTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionMonthPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethod(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusSuccess, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetMonthlyTransactionMethodByMerchantSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Successfully fetched monthly transaction method success stats by merchant",
		Data:    mapMonthlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyTransactionMethodByMerchantSuccess(ctx context.Context, req *pb_transaction.YearMethodTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionYearPaymentmethod, error) {
	data, err := h.repo.GetYearlyMethod(ctx, int(req.GetYear()), transactionrepo.StatusSuccess, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetYearlyTransactionMethodByMerchantSuccess failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Successfully fetched yearly transaction method success stats by merchant",
		Data:    mapYearlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetMonthlyTransactionMethodByMerchantFailed(ctx context.Context, req *pb_transaction.MonthMethodTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionMonthPaymentMethod, error) {
	data, err := h.repo.GetMonthlyMethod(ctx, int(req.GetYear()), int(req.GetMonth()), transactionrepo.StatusFailed, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetMonthlyTransactionMethodByMerchantFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionMonthPaymentMethod{
		Status:  "success",
		Message: "Successfully fetched monthly transaction method failed stats by merchant",
		Data:    mapMonthlyMethod(data),
	}, nil
}

func (h *TransactionStatsHandler) GetYearlyTransactionMethodByMerchantFailed(ctx context.Context, req *pb_transaction.YearMethodTransactionMerchantRequest) (*pb_transaction.ApiResponseTransactionYearPaymentmethod, error) {
	data, err := h.repo.GetYearlyMethod(ctx, int(req.GetYear()), transactionrepo.StatusFailed, req.GetMerchantId())
	if err != nil {
		h.log.Error("GetYearlyTransactionMethodByMerchantFailed failed", zap.Error(err))
		return nil, err
	}
	return &pb_transaction.ApiResponseTransactionYearPaymentmethod{
		Status:  "success",
		Message: "Successfully fetched yearly transaction method failed stats by merchant",
		Data:    mapYearlyMethod(data),
	}, nil
}

// --- Mappers ---

func mapMonthlyAmountSuccess(data []transactionrepo.MonthlyAmount) []*pb_transaction.TransactionMonthlyAmountSuccess {
	var out []*pb_transaction.TransactionMonthlyAmountSuccess
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionMonthlyAmountSuccess{
			Year:         d.Year,
			Month:        d.Month,
			TotalSuccess: int32(d.TotalCount),
			TotalAmount:  int32(d.TotalAmount),
		})
	}
	return out
}

func mapYearlyAmountSuccess(data []transactionrepo.YearlyAmount) []*pb_transaction.TransactionYearlyAmountSuccess {
	var out []*pb_transaction.TransactionYearlyAmountSuccess
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionYearlyAmountSuccess{
			Year:         d.Year,
			TotalSuccess: int32(d.TotalCount),
			TotalAmount:  int32(d.TotalAmount),
		})
	}
	return out
}

func mapMonthlyAmountFailed(data []transactionrepo.MonthlyAmount) []*pb_transaction.TransactionMonthlyAmountFailed {
	var out []*pb_transaction.TransactionMonthlyAmountFailed
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionMonthlyAmountFailed{
			Year:        d.Year,
			Month:       d.Month,
			TotalFailed: int32(d.TotalCount),
			TotalAmount: int32(d.TotalAmount),
		})
	}
	return out
}

func mapYearlyAmountFailed(data []transactionrepo.YearlyAmount) []*pb_transaction.TransactionYearlyAmountFailed {
	var out []*pb_transaction.TransactionYearlyAmountFailed
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionYearlyAmountFailed{
			Year:        d.Year,
			TotalFailed: int32(d.TotalCount),
			TotalAmount: int32(d.TotalAmount),
		})
	}
	return out
}

func mapMonthlyMethod(data []transactionrepo.MonthlyMethod) []*pb_transaction.TransactionMonthlyMethod {
	var out []*pb_transaction.TransactionMonthlyMethod
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionMonthlyMethod{
			Month:             d.Month,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalCount),
			TotalAmount:       int32(d.TotalAmount),
		})
	}
	return out
}

func mapYearlyMethod(data []transactionrepo.YearlyMethod) []*pb_transaction.TransactionYearlyMethod {
	var out []*pb_transaction.TransactionYearlyMethod
	for _, d := range data {
		out = append(out, &pb_transaction.TransactionYearlyMethod{
			Year:              d.Year,
			PaymentMethod:     d.PaymentMethod,
			TotalTransactions: int32(d.TotalCount),
			TotalAmount:       int32(d.TotalAmount),
		})
	}
	return out
}
