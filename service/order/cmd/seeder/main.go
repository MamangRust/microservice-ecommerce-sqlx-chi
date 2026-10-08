package main

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-ecommerce-grpc-order/seeder"
	"github.com/MamangRust/microservice-ecommerce-pkg/database"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// open connects to the database configured via the given DBCluster prefix.
func open(logger logger.LoggerInterface, prefix string) (*sqlx.DB, func(), error) {
	conn, err := database.NewClientWithPrefix(logger, prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to %s: %w", prefix, err)
	}
	closeFn := func() { conn.Close() }
	return conn, closeFn, nil
}

// openItem opens a second handle to the sales database for order_items, kept
// separate from the orders handle so each seeder owns its connection lifecycle.
func openItem(logger logger.LoggerInterface, prefix string) (*sqlx.DB, func(), error) {
	conn, err := database.NewClientWithPrefix(logger, prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("connect to %s: %w", prefix, err)
	}
	closeFn := func() { conn.Close() }
	return conn, closeFn, nil
}

func main() {
	logger, err := logger.NewLogger("seeder", nil)
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}

	if err := dotenv.Viper(); err != nil {
		logger.Fatal("Failed to load .env file", zap.Error(err))
	}

	ctx := context.Background()

	orderDB, closeOrder, err := open(logger, "DB_SALES")
	if err != nil {
		logger.Fatal("Failed to connect to order database", zap.Error(err))
	}
	defer closeOrder()

	orderItemDB, closeItem, err := openItem(logger, "DB_SALES")
	if err != nil {
		logger.Fatal("Failed to connect to sales database", zap.Error(err))
	}
	defer closeItem()

	s := seeder.NewOrderSeeder(orderDB, orderItemDB, ctx, logger)
	if err := s.Seed(); err != nil {
		logger.Fatal("Failed to seed orders", zap.Error(err))
	}

	logger.Info("orders and order_items seeded successfully")
}
