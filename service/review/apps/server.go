package apps

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	userAddr := viper.GetString("GRPC_USER_ADDR")

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	userQueryClient := pb_user.NewUserQueryServiceClient(userConn)

	productAddr := viper.GetString("GRPC_PRODUCT_ADDR")

	productConn, err := grpc.NewClient(productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}
	productQueryClient := pb_product.NewProductQueryServiceClient(productConn)

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.DB,
		userQueryClient,
		productQueryClient,
		repository.GuardOptions{
			User:    []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Product: []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
		})

	obs, _ := observability.NewObservability("review-server", srv.Logger)
	cache := cache.NewMencache(srv.CacheStore)

	svc := service.NewService(&service.Deps{
		Observability: obs,
		Cache:         cache,
		Repositories:  repos,
		Logger:        srv.Logger,
	})

	h := handler.NewHandler(&handler.Deps{Service: svc, Logger: srv.Logger})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_review.RegisterReviewQueryServiceServer(gs, h.ReviewQuery)
		pb_review.RegisterReviewCommandServiceServer(gs, h.ReviewCommand)
	}

	return srv, nil
}
