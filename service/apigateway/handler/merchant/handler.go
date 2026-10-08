package merchanthandler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	merchant_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/merchant"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant"
	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsMerchant struct {
	Client      *grpc.ClientConn
	Router      chi.Router
	Logger      logger.LoggerInterface
	CacheStore  *cache.CacheStore
	UploadImage upload_image.ImageUploads
	ApiHandler  apierror.ApiHandler
}

func RegisterMerchantHandler(deps *DepsMerchant) {
	mapper := apimapper.NewMerchantResponseMapper()
	cache := merchant_cache.NewMerchantMencache(deps.CacheStore)

	handlers := []func(){
		setupMerchantQueryHandler(deps, mapper.QueryMapper(), cache),
		setupMerchantCommandHandler(deps, mapper.CommandMapper(), cache),
	}

	for _, h := range handlers {
		h()
	}
}

func setupMerchantQueryHandler(deps *DepsMerchant, mapper apimapper.MerchantQueryResponseMapper, cache merchant_cache.MerchantQueryCache) func() {
	return func() {
		NewMerchantQueryHandleApi(&merchantQueryHandleDeps{
			client:     pbmerchant.NewMerchantQueryServiceClient(deps.Client),
			router:     deps.Router,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupMerchantCommandHandler(deps *DepsMerchant, mapper apimapper.MerchantCommandResponseMapper, cache merchant_cache.MerchantCommandCache) func() {
	return func() {
		NewMerchantCommandHandleApi(&merchantCommandHandleDeps{
			client:       pbmerchant.NewMerchantCommandServiceClient(deps.Client),
			router:       deps.Router,
			logger:       deps.Logger,
			mapper:       mapper,
			cache:        cache,
			upload_image: deps.UploadImage,
			apiHandler:   deps.ApiHandler,
		})
	}
}
