package bannerhandler

import (
	banner_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/banner"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/banner"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbbanner "github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsBanner struct {
	Client        *grpc.ClientConn
	Router        chi.Router
	Logger        logger.LoggerInterface
	CacheStore    *cache.CacheStore
	Observability observability.TraceLoggerObservability
}

func RegisterBannerHandler(deps *DepsBanner) {
	mapper := apimapper.NewBannerResponseMapper()
	cache := banner_cache.NewBannerMencache(deps.CacheStore)

	NewBannerQueryHandleApi(&bannerQueryHandleDeps{
		client:        pbbanner.NewBannerQueryServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.QueryMapper(),
		cache:         cache,
		observability: deps.Observability,
	})

	NewBannerCommandHandleApi(&bannerCommandHandleDeps{
		client:        pbbanner.NewBannerCommandServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.CommandMapper(),
		cache:         cache,
		observability: deps.Observability,
	})
}
