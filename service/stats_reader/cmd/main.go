package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/handler"
	categoryrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/category"
	orderrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/order"
	transactionrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/transaction"
	"github.com/MamangRust/microservice-ecommerce-pkg/clickhouse"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
	}
	log, _ := logger.NewLogger("stats-reader", nil)

	// The ClickHouse database must exist before NewClient can ping.
	if err := clickhouse.EnsureDatabase(log); err != nil {
		log.Fatal("Failed to ensure ClickHouse database", zap.Error(err))
	}
	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}

	// Guarantee the stats tables exist before serving any query, in case the
	// reader starts before stats-writer has applied the schema.
	if err := clickhouse.ApplySchema(context.Background(), chConn, log); err != nil {
		log.Fatal("Failed to apply ClickHouse schema", zap.Error(err))
	}

	categoryRepo := categoryrepo.NewRepository(chConn)
	orderRepo := orderrepo.NewRepository(chConn)
	transactionRepo := transactionrepo.NewRepository(chConn)

	categoryStatsHandler := handler.NewCategoryStatsHandler(categoryRepo, log)
	orderStatsHandler := handler.NewOrderStatsHandler(orderRepo, log)
	transactionStatsHandler := handler.NewTransactionStatsHandler(transactionRepo, log)

	grpcServer := grpc.NewServer()

	pb_category.RegisterCategoryStatsServiceServer(grpcServer, categoryStatsHandler)
	pb_category.RegisterCategoryStatsByIdServiceServer(grpcServer, categoryStatsHandler)
	pb_category.RegisterCategoryStatsByMerchantServiceServer(grpcServer, categoryStatsHandler)
	pb_order.RegisterOrderStatsServiceServer(grpcServer, orderStatsHandler)
	pb_order.RegisterOrderStatsByMerchantServiceServer(grpcServer, orderStatsHandler)
	pb_transaction.RegisterTransactionStatsServiceServer(grpcServer, transactionStatsHandler)
	pb_transaction.RegisterTransactionStatsByMerchantServiceServer(grpcServer, transactionStatsHandler)

	reflection.Register(grpcServer)

	// The listen address is distinct from GRPC_STATS_READER_ADDR: the reader
	// listens on a bind address (":50070") while apigateway dials a routable
	// host:port (e.g. "stats-reader.ecommerce.svc.cluster.local:50070").
	// STATS_READER_LISTEN_ADDR takes precedence; the dial key is only a
	// convenience fallback for local development (localhost:50070).
	addr := viper.GetString("STATS_READER_LISTEN_ADDR")
	if addr == "" {
		addr = viper.GetString("GRPC_STATS_READER_ADDR")
	}
	if addr == "" {
		addr = ":50070"
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err), zap.String("addr", addr))
	}

	log.Info("Stats Reader starting", zap.String("addr", addr))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Reader...")
	grpcServer.GracefulStop()
}
