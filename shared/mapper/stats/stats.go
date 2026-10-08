package stats

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

// StatsResponseMapper maps the stats-reader (ClickHouse) gRPC responses to the
// REST shapes exposed by the API gateway. The REST shapes are shared with the
// domain services' own stats endpoints (domain/response).
type StatsResponseMapper interface {
	ToApiResponseCategoryMonthlyTotalPrice(res *pb_category.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice
	ToApiResponseCategoryYearlyTotalPrice(res *pb_category.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice
	ToApiResponseCategoryMonthPrice(res *pb_category.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice
	ToApiResponseCategoryYearPrice(res *pb_category.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice

	ToApiResponseOrderMonthlyTotalRevenue(res *pb_order.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue
	ToApiResponseOrderYearlyTotalRevenue(res *pb_order.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue
	ToApiResponseOrderMonthly(res *pb_order.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly
	ToApiResponseOrderYearly(res *pb_order.ApiResponseOrderYearly) *response.ApiResponseOrderYearly

	ToApiResponsesTransactionMonthSuccess(res *pb_transaction.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess
	ToApiResponsesTransactionYearSuccess(res *pb_transaction.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess
	ToApiResponsesTransactionMonthFailed(res *pb_transaction.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed
	ToApiResponsesTransactionYearFailed(res *pb_transaction.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed

	ToApiResponsesTransactionMonthMethod(res *pb_transaction.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod
	ToApiResponsesTransactionYearMethod(res *pb_transaction.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod
}

type statsResponseMapper struct{}

func NewStatsResponseMapper() StatsResponseMapper {
	return &statsResponseMapper{}
}

// --- Category stats ---

func (s *statsResponseMapper) ToApiResponseCategoryMonthlyTotalPrice(res *pb_category.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice {
	data := make([]*response.CategoriesMonthlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesMonthlyTotalPriceResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearlyTotalPrice(res *pb_category.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice {
	data := make([]*response.CategoriesYearlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesYearlyTotalPriceResponse{
			Year:         row.GetYear(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryYearlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryMonthPrice(res *pb_category.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice {
	data := make([]*response.CategoryMonthPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryMonthPriceResponse{
			Month:        row.GetMonth(),
			CategoryID:   int(row.GetCategoryId()),
			CategoryName: row.GetCategoryName(),
			OrderCount:   int(row.GetOrderCount()),
			ItemsSold:    int(row.GetItemsSold()),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearPrice(res *pb_category.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice {
	data := make([]*response.CategoryYearPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryYearPriceResponse{
			Year:               row.GetYear(),
			CategoryID:         int(row.GetCategoryId()),
			CategoryName:       row.GetCategoryName(),
			OrderCount:         int(row.GetOrderCount()),
			ItemsSold:          int(row.GetItemsSold()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseCategoryYearPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

// --- Order stats ---

func (s *statsResponseMapper) ToApiResponseOrderMonthlyTotalRevenue(res *pb_order.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue {
	data := make([]*response.OrderMonthlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyTotalRevenueResponse{
			Year:           row.GetYear(),
			Month:          row.GetMonth(),
			OrderCount:     int(row.GetOrderCount()),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearlyTotalRevenue(res *pb_order.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue {
	data := make([]*response.OrderYearlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyTotalRevenueResponse{
			Year:         row.GetYear(),
			OrderCount:   int(row.GetOrderCount()),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseOrderYearlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderMonthly(res *pb_order.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly {
	data := make([]*response.OrderMonthlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyResponse{
			Month:          row.GetMonth(),
			OrderCount:     int(row.GetOrderCount()),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearly(res *pb_order.ApiResponseOrderYearly) *response.ApiResponseOrderYearly {
	data := make([]*response.OrderYearlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyResponse{
			Year:               row.GetYear(),
			OrderCount:         int(row.GetOrderCount()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			TotalItemsSold:     int(row.GetTotalItemsSold()),
			ActiveCashiers:     int(row.GetActiveCashiers()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseOrderYearly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

// --- Transaction stats ---

func (s *statsResponseMapper) ToApiResponsesTransactionMonthSuccess(res *pb_transaction.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess {
	data := make([]*response.TransactionMonthlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountSuccessResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearSuccess(res *pb_transaction.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess {
	data := make([]*response.TransactionYearlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountSuccessResponse{
			Year:         row.GetYear(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionMonthFailed(res *pb_transaction.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed {
	data := make([]*response.TransactionMonthlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountFailedResponse{
			Year:        row.GetYear(),
			Month:       row.GetMonth(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearFailed(res *pb_transaction.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed {
	data := make([]*response.TransactionYearlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountFailedResponse{
			Year:        row.GetYear(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionMonthMethod(res *pb_transaction.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod {
	data := make([]*response.TransactionMonthlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyMethodResponse{
			Month:             row.GetMonth(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearMethod(res *pb_transaction.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod {
	data := make([]*response.TransactionYearlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyMethodResponse{
			Year:              row.GetYear(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}
