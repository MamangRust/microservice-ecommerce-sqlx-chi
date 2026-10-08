package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// resolveAddr builds the ClickHouse endpoint from CLICKHOUSE_ADDR or
// CLICKHOUSE_HOST/CLICKHOUSE_PORT.
func resolveAddr() string {
	addr := viper.GetString("CLICKHOUSE_ADDR")
	if addr == "" {
		host := viper.GetString("CLICKHOUSE_HOST")
		if host == "" {
			host = "clickhouse"
		}
		port := viper.GetString("CLICKHOUSE_PORT")
		if port == "" {
			port = "9000"
		}
		addr = fmt.Sprintf("%s:%s", host, port)
	}
	return addr
}

// EnsureDatabase creates the configured CLICKHOUSE_DATABASE if it is missing.
// NewClient connects with that database as its default, so the database must
// exist before the client can ping; callers (stats-writer/stats-reader) invoke
// this before NewClient. It is a no-op for the default database.
func EnsureDatabase(l logger.LoggerInterface) error {
	dbName := viper.GetString("CLICKHOUSE_DATABASE")
	if dbName == "" || dbName == "default" {
		return nil
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{resolveAddr()},
		Auth: clickhouse.Auth{
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		return fmt.Errorf("open clickhouse for database bootstrap: %w", err)
	}
	defer conn.Close()

	if err := conn.Exec(context.Background(), "CREATE DATABASE IF NOT EXISTS "+dbName); err != nil {
		l.Error("Failed to create ClickHouse database", zap.Error(err), zap.String("database", dbName))
		return err
	}
	return nil
}

// NewClient opens a ClickHouse connection from env config. It mirrors the
// payment-gateway pkg/clickhouse helper so stats-writer/stats-reader can share
// one bootstrap path. Accepted keys: CLICKHOUSE_ADDR (host:port, takes
// precedence), or CLICKHOUSE_HOST + CLICKHOUSE_PORT; plus CLICKHOUSE_DATABASE,
// CLICKHOUSE_USERNAME and CLICKHOUSE_PASSWORD.
func NewClient(l logger.LoggerInterface) (clickhouse.Conn, error) {
	addr := resolveAddr()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: viper.GetString("CLICKHOUSE_DATABASE"),
			Username: viper.GetString("CLICKHOUSE_USERNAME"),
			Password: viper.GetString("CLICKHOUSE_PASSWORD"),
		},
		DialTimeout: time.Second * 30,
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})

	if err != nil {
		l.Error("Failed to open ClickHouse connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		l.Error("Failed to ping ClickHouse", zap.Error(err))
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	l.Debug("ClickHouse connection established successfully", zap.String("addr", addr))
	return conn, nil
}
