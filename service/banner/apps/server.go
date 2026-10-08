package apps

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-banner/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-banner/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-banner/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-banner/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"google.golang.org/grpc"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	repos := repository.NewRepositories(srv.DB)

	observability, _ := observability.NewObservability("banner-server", srv.Logger)

	cache := cache.NewMencache(srv.CacheStore)

	svc := service.NewService(&service.Deps{
		Cache:         cache,
		Logger:        srv.Logger,
		Repository:    repos,
		Observability: observability,
	})

	h := handler.NewHandler(&handler.Deps{Service: svc, Logger: srv.Logger})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_banner.RegisterBannerQueryServiceServer(gs, h.BannerQuery)
		pb_banner.RegisterBannerCommandServiceServer(gs, h.BannerCommand)
	}

	return srv, nil
}
