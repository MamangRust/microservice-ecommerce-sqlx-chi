package order

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type clickhouseRepository struct {
	conn clickhouse.Conn
}

func (r *clickhouseRepository) GetMonthlyTotalRevenue(ctx context.Context, year, month int, merchantID int32) ([]MonthlyRevenue, error) {
	where := "toYear(created_at) = ? AND toMonth(created_at) = ?"
	args := []interface{}{year, month}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, formatDateTime(created_at, '%%b') AS month, sum(total_price) AS total_revenue
		FROM order_events
		WHERE %s
		GROUP BY year, month, toMonth(created_at)
		ORDER BY year, toMonth(created_at)
	`, where)
	return r.queryMonthlyRevenue(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyTotalRevenue(ctx context.Context, year int, merchantID int32) ([]YearlyRevenue, error) {
	where := "(toYear(created_at) = ? OR toYear(created_at) = ?)"
	args := []interface{}{year, year - 1}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, sum(total_price) AS total_revenue
		FROM order_events
		WHERE %s
		GROUP BY year
		ORDER BY year DESC
	`, where)
	return r.queryYearlyRevenue(ctx, query, args...)
}

func (r *clickhouseRepository) GetMonthlyOrderStats(ctx context.Context, year int, merchantID int32) ([]MonthlyOrder, error) {
	where := "toYear(o.created_at) = ?"
	args := []interface{}{year}
	if merchantID > 0 {
		where += " AND o.merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT
			formatDateTime(o.created_at, '%%b') AS month,
			countDistinct(o.order_id) AS order_count,
			sum(o.total_price) AS total_revenue,
			sum(i.quantity) AS total_items_sold
		FROM order_events o
		LEFT JOIN order_item_events i ON o.order_id = i.order_id
		WHERE %s
		GROUP BY month, toMonth(o.created_at)
		ORDER BY toMonth(o.created_at)
	`, where)
	return r.queryMonthlyOrder(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyOrderStats(ctx context.Context, year int, merchantID int32) ([]YearlyOrder, error) {
	where := "toYear(o.created_at) >= ? AND toYear(o.created_at) <= ?"
	args := []interface{}{year - 4, year}
	if merchantID > 0 {
		where += " AND o.merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT
			toString(toYear(o.created_at)) AS year,
			countDistinct(o.order_id) AS order_count,
			sum(o.total_price) AS total_revenue,
			sum(i.quantity) AS total_items_sold,
			uniqExact(i.product_id) AS unique_products_sold
		FROM order_events o
		LEFT JOIN order_item_events i ON o.order_id = i.order_id
		WHERE %s
		GROUP BY year
		ORDER BY year
	`, where)
	return r.queryYearlyOrder(ctx, query, args...)
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

func (r *clickhouseRepository) queryMonthlyOrder(ctx context.Context, query string, args ...interface{}) ([]MonthlyOrder, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyOrder
	for rows.Next() {
		var m MonthlyOrder
		if err := rows.Scan(&m.Month, &m.OrderCount, &m.TotalRevenue, &m.TotalItemsSold); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryYearlyOrder(ctx context.Context, query string, args ...interface{}) ([]YearlyOrder, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyOrder
	for rows.Next() {
		var y YearlyOrder
		if err := rows.Scan(&y.Year, &y.OrderCount, &y.TotalRevenue, &y.TotalItemsSold, &y.UniqueProductsSold); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, rows.Err()
}
