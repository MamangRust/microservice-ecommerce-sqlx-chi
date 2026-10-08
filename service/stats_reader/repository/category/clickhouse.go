package category

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type clickhouseRepository struct {
	conn clickhouse.Conn
}

// clause turns a Filter into a SQL fragment + args. MerchantID wins when both
// are set so a caller can never produce an ambiguous WHERE.
func (f Filter) clause() (string, []interface{}) {
	switch {
	case f.MerchantID > 0:
		return " AND merchant_id = ?", []interface{}{f.MerchantID}
	case f.CategoryID > 0:
		return " AND category_id = ?", []interface{}{f.CategoryID}
	default:
		return "", nil
	}
}

func (r *clickhouseRepository) GetMonthlyTotalPricing(ctx context.Context, year, month int, filter Filter) ([]MonthlyRevenue, error) {
	where := "toYear(created_at) = ? AND toMonth(created_at) = ?"
	args := []interface{}{year, month}
	clause, clauseArgs := filter.clause()
	where += clause
	args = append(args, clauseArgs...)
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, formatDateTime(created_at, '%%b') AS month, sum(price * quantity) AS total_revenue
		FROM order_item_events
		WHERE %s
		GROUP BY year, month, toMonth(created_at)
		ORDER BY year, toMonth(created_at)
	`, where)
	return r.queryMonthlyRevenue(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyTotalPricing(ctx context.Context, year int, filter Filter) ([]YearlyRevenue, error) {
	where := "(toYear(created_at) = ? OR toYear(created_at) = ?)"
	args := []interface{}{year, year - 1}
	clause, clauseArgs := filter.clause()
	where += clause
	args = append(args, clauseArgs...)
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, sum(price * quantity) AS total_revenue
		FROM order_item_events
		WHERE %s
		GROUP BY year
		ORDER BY year DESC
	`, where)
	return r.queryYearlyRevenue(ctx, query, args...)
}

func (r *clickhouseRepository) GetMonthlyCategoryStats(ctx context.Context, year int, filter Filter) ([]MonthlyCategory, error) {
	where := "toYear(created_at) = ?"
	args := []interface{}{year}
	clause, clauseArgs := filter.clause()
	where += clause
	args = append(args, clauseArgs...)
	query := fmt.Sprintf(`
		SELECT
			formatDateTime(created_at, '%%b') AS month,
			category_id,
			category_name,
			countDistinct(order_id) AS order_count,
			sum(quantity) AS items_sold,
			sum(price * quantity) AS total_revenue
		FROM order_item_events
		WHERE %s
		GROUP BY month, toMonth(created_at), category_id, category_name
		ORDER BY toMonth(created_at), total_revenue DESC
	`, where)
	return r.queryMonthlyCategory(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyCategoryStats(ctx context.Context, year int, filter Filter) ([]YearlyCategory, error) {
	where := "toYear(created_at) >= ? AND toYear(created_at) <= ?"
	args := []interface{}{year - 4, year}
	clause, clauseArgs := filter.clause()
	where += clause
	args = append(args, clauseArgs...)
	query := fmt.Sprintf(`
		SELECT
			toString(toYear(created_at)) AS year,
			category_id,
			category_name,
			countDistinct(order_id) AS order_count,
			sum(quantity) AS items_sold,
			sum(price * quantity) AS total_revenue,
			uniqExact(product_id) AS unique_products_sold
		FROM order_item_events
		WHERE %s
		GROUP BY year, category_id, category_name
		ORDER BY year, total_revenue DESC
	`, where)
	return r.queryYearlyCategory(ctx, query, args...)
}

func (r *clickhouseRepository) queryMonthlyRevenue(ctx context.Context, query string, args ...interface{}) ([]MonthlyRevenue, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyRevenue
	for rows.Next() {
		var m MonthlyRevenue
		if err := rows.Scan(&m.Year, &m.Month, &m.TotalRevenue); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryYearlyRevenue(ctx context.Context, query string, args ...interface{}) ([]YearlyRevenue, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyRevenue
	for rows.Next() {
		var y YearlyRevenue
		if err := rows.Scan(&y.Year, &y.TotalRevenue); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryMonthlyCategory(ctx context.Context, query string, args ...interface{}) ([]MonthlyCategory, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyCategory
	for rows.Next() {
		var m MonthlyCategory
		if err := rows.Scan(&m.Month, &m.CategoryID, &m.CategoryName, &m.OrderCount, &m.ItemsSold, &m.TotalRevenue); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryYearlyCategory(ctx context.Context, query string, args ...interface{}) ([]YearlyCategory, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyCategory
	for rows.Next() {
		var y YearlyCategory
		if err := rows.Scan(&y.Year, &y.CategoryID, &y.CategoryName, &y.OrderCount, &y.ItemsSold, &y.TotalRevenue, &y.UniqueProductsSold); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, rows.Err()
}
