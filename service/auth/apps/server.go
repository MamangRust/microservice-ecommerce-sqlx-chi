package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-ecommerce-auth/cache"
	"github.com/MamangRust/microservice-ecommerce-auth/handler"
	"github.com/MamangRust/microservice-ecommerce-auth/repository"
	"github.com/MamangRust/microservice-ecommerce-auth/service"
	pb_auth "github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	pb_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/auth"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
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

	tokenManager, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		return nil, fmt.Errorf("failed to create token manager: %w", err)
	}

	roleAddr := viper.GetString("GRPC_ROLE_ADDR")

	userAddr := viper.GetString("GRPC_USER_ADDR")

	roleConn, err := grpc.NewClient(
		roleAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(pkgresilience.NewDependencyGuardInterceptor(srv.Logger).UnaryInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to role service: %w", err)
	}

	userConn, err := grpc.NewClient(
		userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(pkgresilience.NewDependencyGuardInterceptor(srv.Logger).UnaryInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	roleQueryClient := pb_role.NewRoleQueryServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)
	userQueryClient := pb_user.NewUserQueryServiceClient(userConn)
	userCommandClient := pb_user.NewUserCommandServiceClient(userConn)

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)
	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, srv.Logger)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, srv.Logger)

	hasher := hash.NewHashingPassword()
	repositories := repository.NewRepositories(&repository.Deps{
		Db:              srv.DB,
		User:            userQueryClient,
		UserCommand:     userCommandClient,
		Role:            roleQueryClient,
		UserRoleCommand: userRoleClient,
		Guards: repository.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	})

	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})

	observability, _ := observability.NewObservability("auth-server", srv.Logger)

	cache := cache.NewMencache(srv.CacheStore)

	outboxService := outbox.NewOutboxService(srv.DB, kafkaOutboxPublisher{k: myKafka}, srv.Logger)

	services := service.NewService(&service.Deps{
		Mencache:      cache,
		Repositories:  repositories,
		Token:         tokenManager,
		Hash:          hasher,
		Logger:        srv.Logger,
		Kafka:         myKafka,
		Pool:          srv.DB,
		Outbox:        outboxService,
		Observability: observability,
	})

	handlers := handler.NewHandler(&handler.Deps{Service: services, Logger: srv.Logger})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_auth.RegisterAuthServiceServer(gs, handlers.Auth)
	}

	// Start the outbox relay so enqueued events are published to Kafka with
	// durable retry and dead-letter semantics.
	go outboxService.Start(srv.Ctx, outbox.OutboxRelayInterval, outbox.OutboxRelayBatchSize)

	return srv, nil
}
