package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	categoryrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/category"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"go.uber.org/zap"
)

// CategoryStatsHandler serves CategoryStatsService, CategoryStatsByIdService
// and CategoryStatsByMerchantService from ClickHouse.
type CategoryStatsHandler struct {
	pb_category.UnimplementedCategoryStatsServiceServer
	pb_category.UnimplementedCategoryStatsByIdServiceServer
	pb_category.UnimplementedCategoryStatsByMerchantServiceServer
	repo categoryrepo.Repository
	log  logger.LoggerInterface
}

func NewCategoryStatsHandler(repo categoryrepo.Repository, log logger.LoggerInterface) *CategoryStatsHandler {
	return &CategoryStatsHandler{repo: repo, log: log}
}

// --- CategoryStatsService ---

func (h *CategoryStatsHandler) FindMonthlyTotalPrices(ctx context.Context, req *pb_category.FindYearMonthTotalPrices) (*pb_category.ApiResponseCategoryMonthlyTotalPrice, error) {
	data, err := h.repo.GetMonthlyTotalPricing(ctx, int(req.GetYear()), int(req.GetMonth()), categoryrepo.Filter{})
	if err != nil {
		h.log.Error("FindMonthlyTotalPrices failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly sales retrieved successfully",
		Data:    mapMonthlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPrices(ctx context.Context, req *pb_category.FindYearTotalPrices) (*pb_category.ApiResponseCategoryYearlyTotalPrice, error) {
	data, err := h.repo.GetYearlyTotalPricing(ctx, int(req.GetYear()), categoryrepo.Filter{})
	if err != nil {
		h.log.Error("FindYearlyTotalPrices failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindMonthPrice(ctx context.Context, req *pb_category.FindYearCategory) (*pb_category.ApiResponseCategoryMonthPrice, error) {
	data, err := h.repo.GetMonthlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{})
	if err != nil {
		h.log.Error("FindMonthPrice failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly payment methods retrieved successfully",
		Data:    mapMonthlyCategory(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearPrice(ctx context.Context, req *pb_category.FindYearCategory) (*pb_category.ApiResponseCategoryYearPrice, error) {
	data, err := h.repo.GetYearlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{})
	if err != nil {
		h.log.Error("FindYearPrice failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyCategory(data),
	}, nil
}

// --- CategoryStatsByIdService ---

func (h *CategoryStatsHandler) FindMonthlyTotalPricesById(ctx context.Context, req *pb_category.FindYearMonthTotalPriceById) (*pb_category.ApiResponseCategoryMonthlyTotalPrice, error) {
	data, err := h.repo.GetMonthlyTotalPricing(ctx, int(req.GetYear()), int(req.GetMonth()), categoryrepo.Filter{CategoryID: req.GetCategoryId()})
	if err != nil {
		h.log.Error("FindMonthlyTotalPricesById failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly sales retrieved successfully",
		Data:    mapMonthlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPricesById(ctx context.Context, req *pb_category.FindYearTotalPriceById) (*pb_category.ApiResponseCategoryYearlyTotalPrice, error) {
	data, err := h.repo.GetYearlyTotalPricing(ctx, int(req.GetYear()), categoryrepo.Filter{CategoryID: req.GetCategoryId()})
	if err != nil {
		h.log.Error("FindYearlyTotalPricesById failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindMonthPriceById(ctx context.Context, req *pb_category.FindYearCategoryById) (*pb_category.ApiResponseCategoryMonthPrice, error) {
	data, err := h.repo.GetMonthlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{CategoryID: req.GetCategoryId()})
	if err != nil {
		h.log.Error("FindMonthPriceById failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly payment methods retrieved successfully",
		Data:    mapMonthlyCategory(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearPriceById(ctx context.Context, req *pb_category.FindYearCategoryById) (*pb_category.ApiResponseCategoryYearPrice, error) {
	data, err := h.repo.GetYearlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{CategoryID: req.GetCategoryId()})
	if err != nil {
		h.log.Error("FindYearPriceById failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyCategory(data),
	}, nil
}

// --- CategoryStatsByMerchantService ---

func (h *CategoryStatsHandler) FindMonthlyTotalPricesByMerchant(ctx context.Context, req *pb_category.FindYearMonthTotalPriceByMerchant) (*pb_category.ApiResponseCategoryMonthlyTotalPrice, error) {
	data, err := h.repo.GetMonthlyTotalPricing(ctx, int(req.GetYear()), int(req.GetMonth()), categoryrepo.Filter{MerchantID: req.GetMerchantId()})
	if err != nil {
		h.log.Error("FindMonthlyTotalPricesByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthlyTotalPrice{
		Status:  "success",
		Message: "Monthly sales retrieved successfully",
		Data:    mapMonthlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearlyTotalPricesByMerchant(ctx context.Context, req *pb_category.FindYearTotalPriceByMerchant) (*pb_category.ApiResponseCategoryYearlyTotalPrice, error) {
	data, err := h.repo.GetYearlyTotalPricing(ctx, int(req.GetYear()), categoryrepo.Filter{MerchantID: req.GetMerchantId()})
	if err != nil {
		h.log.Error("FindYearlyTotalPricesByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearlyTotalPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyPricing(data),
	}, nil
}

func (h *CategoryStatsHandler) FindMonthPriceByMerchant(ctx context.Context, req *pb_category.FindYearCategoryByMerchant) (*pb_category.ApiResponseCategoryMonthPrice, error) {
	data, err := h.repo.GetMonthlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{MerchantID: req.GetMerchantId()})
	if err != nil {
		h.log.Error("FindMonthPriceByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryMonthPrice{
		Status:  "success",
		Message: "Monthly payment methods retrieved successfully",
		Data:    mapMonthlyCategory(data),
	}, nil
}

func (h *CategoryStatsHandler) FindYearPriceByMerchant(ctx context.Context, req *pb_category.FindYearCategoryByMerchant) (*pb_category.ApiResponseCategoryYearPrice, error) {
	data, err := h.repo.GetYearlyCategoryStats(ctx, int(req.GetYear()), categoryrepo.Filter{MerchantID: req.GetMerchantId()})
	if err != nil {
		h.log.Error("FindYearPriceByMerchant failed", zap.Error(err))
		return nil, err
	}
	return &pb_category.ApiResponseCategoryYearPrice{
		Status:  "success",
		Message: "Yearly payment methods retrieved successfully",
		Data:    mapYearlyCategory(data),
	}, nil
}

// --- Mappers ---

func mapMonthlyPricing(data []categoryrepo.MonthlyRevenue) []*pb_category.CategoriesMonthlyTotalPriceResponse {
	var out []*pb_category.CategoriesMonthlyTotalPriceResponse
	for _, d := range data {
		out = append(out, &pb_category.CategoriesMonthlyTotalPriceResponse{
			Year:         d.Year,
			Month:        d.Month,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapYearlyPricing(data []categoryrepo.YearlyRevenue) []*pb_category.CategoriesYearlyTotalPriceResponse {
	var out []*pb_category.CategoriesYearlyTotalPriceResponse
	for _, d := range data {
		out = append(out, &pb_category.CategoriesYearlyTotalPriceResponse{
			Year:         d.Year,
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapMonthlyCategory(data []categoryrepo.MonthlyCategory) []*pb_category.CategoryMonthPriceResponse {
	var out []*pb_category.CategoryMonthPriceResponse
	for _, d := range data {
		out = append(out, &pb_category.CategoryMonthPriceResponse{
			Month:        d.Month,
			CategoryId:   int32(d.CategoryID),
			CategoryName: d.CategoryName,
			OrderCount:   int32(d.OrderCount),
			ItemsSold:    int32(d.ItemsSold),
			TotalRevenue: int32(d.TotalRevenue),
		})
	}
	return out
}

func mapYearlyCategory(data []categoryrepo.YearlyCategory) []*pb_category.CategoryYearPriceResponse {
	var out []*pb_category.CategoryYearPriceResponse
	for _, d := range data {
		out = append(out, &pb_category.CategoryYearPriceResponse{
			Year:               d.Year,
			CategoryId:         int32(d.CategoryID),
			CategoryName:       d.CategoryName,
			OrderCount:         int32(d.OrderCount),
			ItemsSold:          int32(d.ItemsSold),
			TotalRevenue:       int32(d.TotalRevenue),
			UniqueProductsSold: int32(d.UniqueProductsSold),
		})
	}
	return out
}
