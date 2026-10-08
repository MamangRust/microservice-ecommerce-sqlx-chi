package transaction

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type clickhouseRepository struct {
	conn clickhouse.Conn
}

func (r *clickhouseRepository) GetMonthlyAmount(ctx context.Context, year, month int, status Status, merchantID int32) ([]MonthlyAmount, error) {
	where := "toYear(created_at) = ? AND toMonth(created_at) = ? AND status = ?"
	args := []interface{}{year, month, string(status)}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, formatDateTime(created_at, '%%b') AS month, count() AS total_count, sum(amount) AS total_amount
		FROM transaction_events
		WHERE %s
		GROUP BY year, month, toMonth(created_at)
		ORDER BY year, toMonth(created_at)
	`, where)
	return r.queryMonthlyAmount(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyAmount(ctx context.Context, year int, status Status, merchantID int32) ([]YearlyAmount, error) {
	where := "(toYear(created_at) = ? OR toYear(created_at) = ?) AND status = ?"
	args := []interface{}{year, year - 1, string(status)}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, count() AS total_count, sum(amount) AS total_amount
		FROM transaction_events
		WHERE %s
		GROUP BY year
		ORDER BY year DESC
	`, where)
	return r.queryYearlyAmount(ctx, query, args...)
}

func (r *clickhouseRepository) GetMonthlyMethod(ctx context.Context, year, month int, status Status, merchantID int32) ([]MonthlyMethod, error) {
	where := "toYear(created_at) = ? AND toMonth(created_at) = ? AND status = ?"
	args := []interface{}{year, month, string(status)}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT formatDateTime(created_at, '%%b') AS month, payment_method, count() AS total_count, sum(amount) AS total_amount
		FROM transaction_events
		WHERE %s
		GROUP BY month, toMonth(created_at), payment_method
		ORDER BY toMonth(created_at), payment_method
	`, where)
	return r.queryMonthlyMethod(ctx, query, args...)
}

func (r *clickhouseRepository) GetYearlyMethod(ctx context.Context, year int, status Status, merchantID int32) ([]YearlyMethod, error) {
	where := "(toYear(created_at) = ? OR toYear(created_at) = ?) AND status = ?"
	args := []interface{}{year, year - 1, string(status)}
	if merchantID > 0 {
		where += " AND merchant_id = ?"
		args = append(args, merchantID)
	}
	query := fmt.Sprintf(`
		SELECT toString(toYear(created_at)) AS year, payment_method, count() AS total_count, sum(amount) AS total_amount
		FROM transaction_events
		WHERE %s
		GROUP BY year, payment_method
		ORDER BY year DESC, payment_method
	`, where)
	return r.queryYearlyMethod(ctx, query, args...)
}

func (r *clickhouseRepository) queryMonthlyAmount(ctx context.Context, query string, args ...interface{}) ([]MonthlyAmount, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyAmount
	for rows.Next() {
		var m MonthlyAmount
		if err := rows.Scan(&m.Year, &m.Month, &m.TotalCount, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryYearlyAmount(ctx context.Context, query string, args ...interface{}) ([]YearlyAmount, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyAmount
	for rows.Next() {
		var y YearlyAmount
		if err := rows.Scan(&y.Year, &y.TotalCount, &y.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, y)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryMonthlyMethod(ctx context.Context, query string, args ...interface{}) ([]MonthlyMethod, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []MonthlyMethod
	for rows.Next() {
		var m MonthlyMethod
		if err := rows.Scan(&m.Month, &m.PaymentMethod, &m.TotalCount, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *clickhouseRepository) queryYearlyMethod(ctx context.Context, query string, args ...interface{}) ([]YearlyMethod, error) {
	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []YearlyMethod
	for rows.Next() {
		var m YearlyMethod
		if err := rows.Scan(&m.Year, &m.PaymentMethod, &m.TotalCount, &m.TotalAmount); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}
