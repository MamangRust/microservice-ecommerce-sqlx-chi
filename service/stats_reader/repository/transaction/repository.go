// Package transaction reads transaction statistics from ClickHouse. It
// replaces the transaction slice of the old single stats-reader Repository god
// interface, and the old stringly-typed status parameter with a typed Status.
package transaction

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Status is the payment status a transaction aggregation is filtered by.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

// Row shapes returned by the ClickHouse queries. Amounts are int64 because
// ClickHouse SUM returns a 64-bit integer for Int64 columns.
type MonthlyAmount struct {
	Year        string
	Month       string
	TotalCount  uint64
	TotalAmount int64
}

type YearlyAmount struct {
	Year        string
	TotalCount  uint64
	TotalAmount int64
}

type MonthlyMethod struct {
	Month         string
	PaymentMethod string
	TotalCount    uint64
	TotalAmount   int64
}

type YearlyMethod struct {
	Year          string
	PaymentMethod string
	TotalCount    uint64
	TotalAmount   int64
}

// Repository reads transaction statistics (transaction_events). merchantID <= 0
// aggregates every merchant; a positive merchantID scopes to that merchant.
type Repository interface {
	GetMonthlyAmount(ctx context.Context, year, month int, status Status, merchantID int32) ([]MonthlyAmount, error)
	GetYearlyAmount(ctx context.Context, year int, status Status, merchantID int32) ([]YearlyAmount, error)
	GetMonthlyMethod(ctx context.Context, year, month int, status Status, merchantID int32) ([]MonthlyMethod, error)
	GetYearlyMethod(ctx context.Context, year int, status Status, merchantID int32) ([]YearlyMethod, error)
}

// NewRepository returns the ClickHouse-backed transaction repository.
func NewRepository(conn clickhouse.Conn) Repository {
	return &clickhouseRepository{conn: conn}
}
