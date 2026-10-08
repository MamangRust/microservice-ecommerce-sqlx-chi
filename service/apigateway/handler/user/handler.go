package userhandler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	user_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/user"
	pbuser "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsUser struct {
	Client *grpc.ClientConn

	Router chi.Router

	Logger logger.LoggerInterface

	Cache *cache.CacheStore

	ApiHandler apierror.ApiHandler
}

func RegisterUserHandler(deps *DepsUser) {
	mapper := apimapper.NewUserResponseMapper()

	cache := user_cache.NewUserMencache(deps.Cache)

	handlers := []func(){
		setupUserQueryHandler(deps, mapper.QueryMapper(), cache),
		setupUserCommandHandler(deps, mapper.CommandMapper(), cache),
	}

	for _, h := range handlers {
		h()
	}
}

func setupUserQueryHandler(deps *DepsUser, mapper apimapper.UserQueryResponseMapper, cache user_cache.UserMencache) func() {
	return func() {
		NewUserQueryHandleApi(&userQueryHandleDeps{
			client:     pbuser.NewUserQueryServiceClient(deps.Client),
			router:     deps.Router,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}

func setupUserCommandHandler(deps *DepsUser, mapper apimapper.UserCommandResponseMapper, cache user_cache.UserMencache) func() {
	return func() {
		NewUserCommandHandleApi(&userCommandHandleDeps{
			client:     pbuser.NewUserCommandServiceClient(deps.Client),
			router:     deps.Router,
			logger:     deps.Logger,
			mapper:     mapper,
			cache:      cache,
			apiHandler: deps.ApiHandler,
		})
	}
}
