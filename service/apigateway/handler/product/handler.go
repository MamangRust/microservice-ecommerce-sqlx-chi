package producthandler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	product_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/product"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/product"
	pbproduct "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsProduct struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
	Upload     upload_image.ImageUploads
	ApiHandler apierror.ApiHandler
}

func RegisterProductHandler(deps *DepsProduct) {
	mapper := apimapper.NewProductResponseMapper()
	cache := product_cache.NewProductMencache(deps.CacheStore)

	queryClient := pbproduct.NewProductQueryServiceClient(deps.Client)

	NewProductQueryHandleApi(&productQueryHandleDeps{
		client:     queryClient,
		router:     deps.Router,
		logger:     deps.Logger,
		mapper:     mapper.QueryMapper(),
		cache:      cache,
		apiHandler: deps.ApiHandler,
	})

	NewProductCommandHandleApi(&productCommandHandleDeps{
		client:       pbproduct.NewProductCommandServiceClient(deps.Client),
		router:       deps.Router,
		logger:       deps.Logger,
		mapper:       mapper.CommandMapper(),
		cache:        cache,
		upload_image: deps.Upload,
		apiHandler:   deps.ApiHandler,
	})

}
