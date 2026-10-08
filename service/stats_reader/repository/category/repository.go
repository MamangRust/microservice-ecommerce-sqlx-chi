// Package category reads category pricing statistics from ClickHouse. It
// replaces the category slice of the old single stats-reader Repository god
// interface, and the old stringly-typed filterField ("", "merchant_id",
// "category_id") with a typed Filter.
package category

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Row shapes returned by the ClickHouse queries. Amounts are int64 because
// ClickHouse SUM returns a 64-bit integer for Int64 columns.
type MonthlyRevenue struct {
	Year         string
	Month        string
	TotalRevenue int64
}

type YearlyRevenue struct {
	Year         string
	TotalRevenue int64
}

type MonthlyCategory struct {
	Month        string
	CategoryID   uint64
	CategoryName string
	OrderCount   uint64
	ItemsSold    uint64
	TotalRevenue int64
}

type YearlyCategory struct {
	Year               string
	CategoryID         uint64
	CategoryName       string
	OrderCount         uint64
	ItemsSold          uint64
	TotalRevenue       int64
	UniqueProductsSold uint64
}

// Filter scopes a category aggregation. The zero value aggregates every
// category; set MerchantID or CategoryID (not both) to narrow it. This is the
// typed replacement for the old filterField string and mirrors the main /
// ByMerchant / ById service split.
type Filter struct {
	MerchantID int32
	CategoryID int32
}

// Repository reads category pricing statistics (order_item_events).
type Repository interface {
	GetMonthlyTotalPricing(ctx context.Context, year, month int, filter Filter) ([]MonthlyRevenue, error)
	GetYearlyTotalPricing(ctx context.Context, year int, filter Filter) ([]YearlyRevenue, error)
	GetMonthlyCategoryStats(ctx context.Context, year int, filter Filter) ([]MonthlyCategory, error)
	GetYearlyCategoryStats(ctx context.Context, year int, filter Filter) ([]YearlyCategory, error)
}

// NewRepository returns the ClickHouse-backed category repository.
func NewRepository(conn clickhouse.Conn) Repository {
	return &clickhouseRepository{conn: conn}
}
