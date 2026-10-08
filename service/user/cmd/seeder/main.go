package main

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-ecommerce-grpc-user/seeder"
	"github.com/MamangRust/microservice-ecommerce-pkg/database"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
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

func main() {
	logger, err := logger.NewLogger("seeder", nil)
	if err != nil {
		logger.Fatal("Failed to initialize logger", zap.Error(err))
	}

	if err := dotenv.Viper(); err != nil {
		logger.Fatal("Failed to load .env file", zap.Error(err))
	}

	ctx := context.Background()

	q, closeFn, err := open(logger, "DB_IDENTITY")
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer closeFn()

	s := seeder.NewUserSeeder(q, hash.NewHashingPassword(), ctx, logger)
	if err := s.Seed(); err != nil {
		logger.Fatal("Failed to seed users", zap.Error(err))
	}

	logger.Info("users seeded successfully")
}
