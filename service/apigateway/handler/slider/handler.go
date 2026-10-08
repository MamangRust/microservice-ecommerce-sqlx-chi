package sliderhandler

import (
	slider_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/slider"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/slider"
	pbslider "github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsSlider struct {
	Client *grpc.ClientConn
	Router chi.Router
	Logger logger.LoggerInterface
	Cache  *cache.CacheStore
	Upload upload_image.ImageUploads
}

func RegisterSliderHandler(deps *DepsSlider) {
	mapper := apimapper.NewSliderResponseMapper()
	cache := slider_cache.NewSliderMencache(deps.Cache)

	NewSliderQueryHandleApi(&sliderQueryHandleDeps{
		client: pbslider.NewSliderQueryServiceClient(deps.Client),
		router: deps.Router,
		logger: deps.Logger,
		mapper: mapper.QueryMapper(),
		cache:  cache.QueryCache(),
	})

	NewSliderCommandHandleApi(&sliderCommandHandleDeps{
		client:      pbslider.NewSliderCommandServiceClient(deps.Client),
		router:      deps.Router,
		logger:      deps.Logger,
		mapper:      mapper.CommandMapper(),
		queryMapper: mapper.QueryMapper(),
		cache:       cache.CommandCache(),
		upload:      deps.Upload,
	})
}
