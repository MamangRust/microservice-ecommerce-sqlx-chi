package apps

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/service"
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
	mencache := cache.NewMencache(srv.CacheStore)
	obs, _ := observability.NewObservability("shipping_address-server", srv.Logger)

	svc := service.NewService(&service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        srv.Logger,
		Observability: obs,
	})

	h := handler.NewHandler(&handler.Deps{
		Service: svc,
		Logger:  srv.Logger,
	})

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_shipping_address.RegisterShippingQueryServiceServer(gs, h.ShippingQuery)
		pb_shipping_address.RegisterShippingCommandServiceServer(gs, h.ShippingCommand)
	}

	return srv, nil
}
