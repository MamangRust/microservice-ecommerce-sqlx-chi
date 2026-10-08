package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/handler"
	merchantKafka "github.com/MamangRust/microservice-ecommerce-grpc-merchant/kafka"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/service"
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	pb_merchant_document "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/outbox"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pkgresilience "github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"google.golang.org/grpc/credentials/insecure"
)

// kafkaOutboxPublisher adapts the ecommerce *kafka.Kafka (whose SendMessage
// takes no context) to the outbox.OutboxPublisher contract.
type kafkaOutboxPublisher struct {
	k *kafka.Kafka
}

func (p kafkaOutboxPublisher) SendMessage(_ context.Context, topic, key string, value []byte) error {
	return p.k.SendMessage(topic, key, value)
}

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	userAddr := viper.GetString("GRPC_USER_ADDR")

	userConn, err := grpc.NewClient(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(pkgresilience.NewDependencyGuardInterceptor(srv.Logger).UnaryInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	userQueryClient := pb_user.NewUserQueryServiceClient(userConn)

	repos := repository.NewRepositories(srv.DB, userQueryClient,
		repository.GuardOptions{
			User: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	)

	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})
	mencache := cache.NewMencache(srv.CacheStore)
	obs, _ := observability.NewObservability(viper.GetString("merchant-server"), srv.Logger)

	outboxService := outbox.NewOutboxService(srv.DB, kafkaOutboxPublisher{k: myKafka}, srv.Logger)

	svc := service.NewService(&service.Deps{
		Kafka:         myKafka,
		Repositories:  repos,
		Mencache:      mencache,
		Pool:          srv.DB,
		Outbox:        outboxService,
		Logger:        srv.Logger,
		Observability: obs,
	})

	go outboxService.Start(srv.Ctx, outbox.OutboxRelayInterval, outbox.OutboxRelayBatchSize)

	h := handler.NewHandler(&handler.Deps{
		Service: svc,
		Logger:  srv.Logger,
	})

	if err := myKafka.StartConsumersWithContext(srv.Ctx, []string{"merchant-service-topic-transaction-event"}, "merchant-service-group", merchantKafka.NewTransactionConsumer(srv.Ctx, mencache.MerchantCommandCache, srv.Logger)); err != nil {
		return nil, fmt.Errorf("failed to start merchant transaction consumer: %w", err)
	}

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_merchant.RegisterMerchantQueryServiceServer(gs, h.MerchantQuery)
		pb_merchant.RegisterMerchantCommandServiceServer(gs, h.MerchantCommandHandler)
		pb_merchant_document.RegisterMerchantDocumentQueryServiceServer(gs, h.MerchantDocumentQuery)
		pb_merchant_document.RegisterMerchantDocumentCommandServiceServer(gs, h.MerchantDocumentCommand)
	}

	return srv, nil
}
