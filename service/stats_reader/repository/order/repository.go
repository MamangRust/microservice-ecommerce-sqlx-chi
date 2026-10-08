// Package order reads order statistics (revenue + order aggregates) from
// ClickHouse. It replaces the order slice of the old single stats-reader
// Repository god interface.
package order

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

type MonthlyOrder struct {
	Month          string
	OrderCount     uint64
	TotalRevenue   int64
	TotalItemsSold uint64
}

type YearlyOrder struct {
	Year               string
	OrderCount         uint64
	TotalRevenue       int64
	TotalItemsSold     uint64
	UniqueProductsSold uint64
}

// Repository reads order statistics. merchantID <= 0 aggregates every
// merchant (the main service); a positive merchantID scopes to that merchant
// (the ByMerchant service).
type Repository interface {
	GetMonthlyTotalRevenue(ctx context.Context, year, month int, merchantID int32) ([]MonthlyRevenue, error)
	GetYearlyTotalRevenue(ctx context.Context, year int, merchantID int32) ([]YearlyRevenue, error)
	GetMonthlyOrderStats(ctx context.Context, year int, merchantID int32) ([]MonthlyOrder, error)
	GetYearlyOrderStats(ctx context.Context, year int, merchantID int32) ([]YearlyOrder, error)
}

// NewRepository returns the ClickHouse-backed order repository.
func NewRepository(conn clickhouse.Conn) Repository {
	return &clickhouseRepository{conn: conn}
}
