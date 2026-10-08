package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	categoryadapter "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	orderadapter "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	productadapter "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	transactionadapter "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	catadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/category"
	orderitemadapterpkg "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	orderadapterpkg "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order"
	productadapterpkg "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	txnadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/backfill"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/usecase"
	"github.com/MamangRust/microservice-ecommerce-pkg/clickhouse"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		zap.L().Error("Failed to load configuration", zap.Error(err))
	}
	log, _ := logger.NewLogger("stats-writer", nil)

	// The ClickHouse database must exist before NewClient can ping (it connects
	// with the configured database as default), so create it first.
	if err := clickhouse.EnsureDatabase(log); err != nil {
		log.Fatal("Failed to ensure ClickHouse database", zap.Error(err))
	}
	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}

	// Guarantee the stats tables exist before any read/write.
	if err := clickhouse.ApplySchema(context.Background(), chConn, log); err != nil {
		log.Fatal("Failed to apply ClickHouse schema", zap.Error(err))
	}

	repo := repository.NewClickhouseRepository(chConn, log)
	uc := usecase.NewStatsUseCase(repo)

	// `stats-writer backfill` materializes historical OLTP rows into
	// ClickHouse and exits. Everything else runs the live Kafka consumer.
	if len(os.Args) > 1 && os.Args[1] == "backfill" {
		bf, err := newBackfiller(log, repo)
		if err != nil {
			log.Fatal("Failed to build backfiller", zap.Error(err))
		}
		if err := bf.Run(context.Background()); err != nil {
			log.Fatal("Backfill failed", zap.Error(err))
		}
		log.Info("backfill finished")
		return
	}

	brokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")
	if len(brokers) == 0 || brokers[0] == "" {
		brokers = []string{"kafka:9092", "localhost:9092"}
	}
	k := kafka.NewKafka(log, brokers)

	statsHandler := handler.NewStatsHandler(uc, log)
	if err := k.StartConsumers(handler.StatsTopics(), "ecommerce-stats-writer", statsHandler); err != nil {
		log.Fatal("Failed to start Kafka consumers", zap.Error(err))
	}
	log.Info("Stats Writer consuming", zap.Strings("topics", handler.StatsTopics()), zap.Strings("brokers", brokers))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Writer...")
	if err := uc.Close(); err != nil {
		log.Error("Failed to close stats usecase", zap.Error(err))
	}
}

// newBackfiller dials the owning services whose OLTP data the backfill
// materializes and wires them into a Backfiller. The addresses come from the
// same GRPC_<SVC>_ADDR config the live services use.
func newBackfiller(log logger.LoggerInterface, repo repository.Repository) (*backfill.Backfiller, error) {
	dial := func(addrKey string) (*grpc.ClientConn, error) {
		addr := viper.GetString(addrKey)
		if addr == "" {
			return nil, fmt.Errorf("missing %s in configuration", addrKey)
		}
		return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	orderConn, err := dial("GRPC_ORDER_ADDR")
	if err != nil {
		return nil, err
	}
	orderItemConn, err := dial("GRPC_ORDER_ITEM_ADDR")
	if err != nil {
		return nil, err
	}
	productConn, err := dial("GRPC_PRODUCT_ADDR")
	if err != nil {
		return nil, err
	}
	categoryConn, err := dial("GRPC_CATEGORY_ADDR")
	if err != nil {
		return nil, err
	}
	transactionConn, err := dial("GRPC_TRANSACTION_ADDR")
	if err != nil {
		return nil, err
	}

	orderRepo := orderadapterpkg.NewBulkAdapter(orderadapter.NewOrderQueryServiceClient(orderConn))
	orderItemRepo := orderitemadapterpkg.NewBulkAdapter(orderitemadapter.NewOrderItemQueryServiceClient(orderItemConn))
	productRepo := productadapterpkg.NewAdapter(
		productadapter.NewProductQueryServiceClient(productConn),
		productadapter.NewProductCommandServiceClient(productConn),
	)
	categoryRepo := catadapter.NewBulkAdapter(categoryadapter.NewCategoryQueryServiceClient(categoryConn))
	transactionRepo := txnadapter.NewBulkAdapter(transactionadapter.NewTransactionQueryServiceClient(transactionConn))

	return backfill.New(log, repo, orderRepo, orderItemRepo, productRepo, categoryRepo, transactionRepo), nil
}
