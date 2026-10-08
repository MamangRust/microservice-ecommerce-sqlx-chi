package transactionhandler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	transaction_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/transaction"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/transaction"
	pbtransaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsTransaction struct {
	Client     *grpc.ClientConn
	Router     chi.Router
	Logger     logger.LoggerInterface
	CacheStore *cache.CacheStore
	ApiHandler apierror.ApiHandler
}

func RegisterTransactionHandler(deps *DepsTransaction) {
	mapper := apimapper.NewTransactionResponseMapper()
	cache := transaction_cache.NewTransactionMencache(deps.CacheStore)

	queryClient := pbtransaction.NewTransactionQueryServiceClient(deps.Client)
	commandClient := pbtransaction.NewTransactionCommandServiceClient(deps.Client)

	NewTransactionQueryHandleApi(&transactionQueryHandleDeps{
		queryClient: queryClient,
		router:      deps.Router,
		logger:      deps.Logger,
		mapper:      mapper.QueryMapper(),
		cache:       cache,
		apiHandler:  deps.ApiHandler,
	})

	NewTransactionCommandHandleApi(&transactionCommandHandleDeps{
		client:     commandClient,
		router:     deps.Router,
		logger:     deps.Logger,
		mapper:     mapper.CommandMapper(),
		cache:      cache,
		apiHandler: deps.ApiHandler,
	})
}
