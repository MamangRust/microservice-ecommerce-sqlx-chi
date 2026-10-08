package categoryhandler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	category_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/category"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/category"
	pbcategory "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsCategory struct {
	Client      *grpc.ClientConn
	Router      chi.Router
	Logger      logger.LoggerInterface
	CacheStore  *cache.CacheStore
	UploadImage upload_image.ImageUploads
	ApiHandler  apierror.ApiHandler
}

func RegisterCategoryHandler(deps *DepsCategory) {
	mapper := apimapper.NewCategoryResponseMapper()
	cache := category_cache.NewCategoryMencache(deps.CacheStore)

	handlers := []func(){
		setupCategoryQueryHandler(deps, mapper.QueryMapper(), cache),
		setupCategoryCommandHandler(deps, mapper.CommandMapper(), cache),
	}

	for _, h := range handlers {
		h()
	}
}

func setupCategoryQueryHandler(deps *DepsCategory, mapper apimapper.CategoryQueryResponseMapper, cache category_cache.CategoryMencache) func() {
	return func() {
		NewCategoryQueryHandleApi(&categoryQueryHandleDeps{
			client:     pbcategory.NewCategoryQueryServiceClient(deps.Client),
			router:     deps.Router,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupCategoryCommandHandler(deps *DepsCategory, mapper apimapper.CategoryCommandResponseMapper, cache category_cache.CategoryMencache) func() {
	return func() {
		NewCategoryCommandHandleApi(&categoryCommandHandleDeps{
			client:       pbcategory.NewCategoryCommandServiceClient(deps.Client),
			router:       deps.Router,
			logger:       deps.Logger,
			mapper:       mapper,
			cache:        cache,
			upload_image: deps.UploadImage,
			apiHandler:   deps.ApiHandler,
		})
	}
}
