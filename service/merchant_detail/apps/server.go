package apps

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/service"
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	pb_merchant_detail "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	pb_merchant_social_link "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_social_link"
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

	merchantAddr := viper.GetString("GRPC_MERCHANT_ADDR")

	merchantConn, err := grpc.NewClient(
		merchantAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to merchant service: %w", err)
	}

	merchantQueryClient := pb_merchant.NewMerchantQueryServiceClient(merchantConn)

	repos := repository.NewRepositories(srv.DB, merchantQueryClient,
		repository.GuardOptions{
			Merchant: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	)

	obs, _ := observability.NewObservability("merchant-detail-server", srv.Logger)

	cache := cache.NewMencache(srv.CacheStore)

	svc := service.NewService(&service.Deps{
		Cache:         cache,
		Logger:        srv.Logger,
		Repository:    repos,
		Observability: obs,
	})

	h := handler.NewHandler(&handler.Deps{Service: svc, Logger: srv.Logger})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_merchant_detail.RegisterMerchantDetailQueryServiceServer(gs, h.MerchantDetailQuery)
		pb_merchant_detail.RegisterMerchantDetailCommandServiceServer(gs, h.MerchantDetailCommand)
		pb_merchant_social_link.RegisterMerchantSocialCommandServiceServer(gs, h.MerchantSocialLinkCommand)
	}

	return srv, nil
}
