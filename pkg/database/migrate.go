package database

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// RunMigrations executes database migrations using goose against the DB
// configured via the given prefix (e.g. "DB_IDENTITY"). Each bounded context
// owns its own PostgreSQL instance, so the prefix resolves its own
// <Prefix>_HOST / _PORT / _USERNAME / _NAME / _PASSWORD keys. Only the username
// and password fall back to the generic DB_* keys; a missing host, port or
// database name fails fast instead of silently targeting another context.
// path: directory containing migration files.
func RunMigrations(log logger.LoggerInterface, prefix, path string) error {
	if prefix == "" {
		prefix = "DB"
	}

	host := viper.GetString(fmt.Sprintf("%s_HOST", prefix))
	port := viper.GetString(fmt.Sprintf("%s_PORT", prefix))
	user := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_USERNAME", prefix)), viper.GetString("DB_USERNAME"))
	dbname := viper.GetString(fmt.Sprintf("%s_NAME", prefix))
	password := firstNonEmpty(viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix)), viper.GetString("DB_PASSWORD"))

	if host == "" || port == "" || dbname == "" {
		log.Error("Incomplete database configuration for migrations",
			zap.String("cluster", prefix),
			zap.String("host", host),
			zap.String("port", port),
			zap.String("dbname", dbname),
		)
		return fmt.Errorf("incomplete database configuration for %s (host=%q port=%q dbname=%q)",
			prefix, host, port, dbname)
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password,
	)

	db, err := goose.OpenDBWithDriver("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database after migrations", zap.Error(err))
		}
	}()

	log.Info("Running database migrations",
		zap.String("path", path),
		zap.String("dbname", dbname),
		zap.String("prefix", prefix),
	)

	if err := goose.RunContext(context.Background(), "up", db, path); err != nil {
		return fmt.Errorf("migration 'up' failed: %w", err)
	}

	log.Info("Database migrations completed successfully")
	return nil
}
