package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	orderrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/order"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"go.uber.org/zap"
)

// OrderStatsHandler serves OrderStatsService (revenue + order aggregates) from
// ClickHouse.
type OrderStatsHandler struct {
	pb_order.UnimplementedOrderStatsServiceServer
	pb_order.UnimplementedOrderStatsByMerchantServiceServer
	repo orderrepo.Repository
	log  logger.LoggerInterface
}

func NewOrderStatsHandler(repo orderrepo.Repository, log logger.LoggerInterface) *OrderStatsHandler {
	return &OrderStatsHandler{repo: repo, log: log}
}

func (h *OrderStatsHandler) FindMonthlyTotalRevenue(ctx context.Context, req *pb_order.FindYearMonthTotalRevenue) (*pb_order.ApiResponseOrderMonthlyTotalRevenue, error) {
	data, err := h.repo.GetMonthlyTotalRevenue(ctx, int(req.GetYear()), int(req.GetMonth()), 0)
	if err != nil {
		h.log.Error("FindMonthlyTotalRevenue failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderMonthlyTotalRevenue{
		Status:  "success",
		Message: "Monthly sales retrieved successfully",
		Data:    mapMonthlyRevenue(data),
	}, nil
}

func (h *OrderStatsHandler) FindYearlyTotalRevenue(ctx context.Context, req *pb_order.FindYearTotalRevenue) (*pb_order.ApiResponseOrderYearlyTotalRevenue, error) {
	data, err := h.repo.GetYearlyTotalRevenue(ctx, int(req.GetYear()), 0)
	if err != nil {
		h.log.Error("FindYearlyTotalRevenue failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderYearlyTotalRevenue{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyRevenue(data),
	}, nil
}

func (h *OrderStatsHandler) FindMonthlyRevenue(ctx context.Context, req *pb_order.FindYearOrder) (*pb_order.ApiResponseOrderMonthly, error) {
	data, err := h.repo.GetMonthlyOrderStats(ctx, int(req.GetYear()), 0)
	if err != nil {
		h.log.Error("FindMonthlyRevenue failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderMonthly{
		Status:  "success",
		Message: "Monthly revenue data retrieved",
		Data:    mapMonthlyOrder(data),
	}, nil
}

func (h *OrderStatsHandler) FindYearlyRevenue(ctx context.Context, req *pb_order.FindYearOrder) (*pb_order.ApiResponseOrderYearly, error) {
	data, err := h.repo.GetYearlyOrderStats(ctx, int(req.GetYear()), 0)
	if err != nil {
		h.log.Error("FindYearlyRevenue failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderYearly{
		Status:  "success",
		Message: "Yearly revenue data retrieved",
		Data:    mapYearlyOrder(data),
	}, nil
}

func (h *OrderStatsHandler) FindMonthlyTotalRevenueByMerchant(ctx context.Context, req *pb_order.FindYearMonthTotalRevenueByMerchant) (*pb_order.ApiResponseOrderMonthlyTotalRevenue, error) {
	data, err := h.repo.GetMonthlyTotalRevenue(ctx, int(req.GetYear()), int(req.GetMonth()), req.GetMerchantId())
	if err != nil {
		h.log.Error("FindMonthlyTotalRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderMonthlyTotalRevenue{
		Status:  "success",
		Message: "Monthly sales retrieved successfully",
		Data:    mapMonthlyRevenue(data),
	}, nil
}

func (h *OrderStatsHandler) FindYearlyTotalRevenueByMerchant(ctx context.Context, req *pb_order.FindYearTotalRevenueByMerchant) (*pb_order.ApiResponseOrderYearlyTotalRevenue, error) {
	data, err := h.repo.GetYearlyTotalRevenue(ctx, int(req.GetYear()), req.GetMerchantId())
	if err != nil {
		h.log.Error("FindYearlyTotalRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderYearlyTotalRevenue{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyRevenue(data),
	}, nil
}

func (h *OrderStatsHandler) FindMonthlyRevenueByMerchant(ctx context.Context, req *pb_order.FindYearOrderByMerchant) (*pb_order.ApiResponseOrderMonthly, error) {
	data, err := h.repo.GetMonthlyOrderStats(ctx, int(req.GetYear()), req.GetMerchantId())
	if err != nil {
		h.log.Error("FindMonthlyRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderMonthly{
		Status:  "success",
		Message: "Monthly revenue by merchant data retrieved",
		Data:    mapMonthlyOrder(data),
	}, nil
}

func (h *OrderStatsHandler) FindYearlyRevenueByMerchant(ctx context.Context, req *pb_order.FindYearOrderByMerchant) (*pb_order.ApiResponseOrderYearly, error) {
	data, err := h.repo.GetYearlyOrderStats(ctx, int(req.GetYear()), req.GetMerchantId())
	if err != nil {
		h.log.Error("FindYearlyRevenueByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_order.ApiResponseOrderYearly{
		Status:  "success",
		Message: "Yearly revenue by merchant data retrieved",
		Data:    mapYearlyOrder(data),
	}, nil
}

// --- Mappers ---

func mapMonthlyRevenue(data []orderrepo.MonthlyRevenue) []*pb_order.OrderMonthlyTotalRevenueResponse {
	var out []*pb_order.OrderMonthlyTotalRevenueResponse
	for _, d := range data {
		out = append(out, &pb_order.OrderMonthlyTotalRevenueResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapYearlyRevenue(data []orderrepo.YearlyRevenue) []*pb_order.OrderYearlyTotalRevenueResponse {
	var out []*pb_order.OrderYearlyTotalRevenueResponse
	for _, d := range data {
		out = append(out, &pb_order.OrderYearlyTotalRevenueResponse{
			Year:         d.Year,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapMonthlyOrder(data []orderrepo.MonthlyOrder) []*pb_order.OrderMonthlyResponse {
	var out []*pb_order.OrderMonthlyResponse
	for _, d := range data {
		out = append(out, &pb_order.OrderMonthlyResponse{
			Month:          d.Month,
			OrderCount:     int32(d.OrderCount),
			TotalRevenue:   int32(d.TotalRevenue),
			TotalItemsSold: int32(d.TotalItemsSold),
		})
	}
	return out
}

func mapYearlyOrder(data []orderrepo.YearlyOrder) []*pb_order.OrderYearlyResponse {
	var out []*pb_order.OrderYearlyResponse
	for _, d := range data {
		out = append(out, &pb_order.OrderYearlyResponse{
			Year:               d.Year,
			OrderCount:         int32(d.OrderCount),
			TotalRevenue:       int32(d.TotalRevenue),
			TotalItemsSold:     int32(d.TotalItemsSold),
			UniqueProductsSold: int32(d.UniqueProductsSold),
		})
	}
	return out
}
