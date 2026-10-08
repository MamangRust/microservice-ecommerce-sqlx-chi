package reviewhandler

import (
	review_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/review"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbreview "github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsReview struct {
	Client        *grpc.ClientConn
	Router        chi.Router
	Logger        logger.LoggerInterface
	Cache         *cache.CacheStore
	Observability observability.TraceLoggerObservability
}

func RegisterReviewHandler(deps *DepsReview) {
	mapper := apimapper.NewReviewResponseMapper()
	cache := review_cache.NewReviewMencache(deps.Cache)

	NewReviewQueryHandleApi(&reviewQueryHandleDeps{
		client:        pbreview.NewReviewQueryServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.QueryMapper(),
		cache:         cache.QueryCache(),
		observability: deps.Observability,
	})

	NewReviewCommandHandleApi(&reviewCommandHandleDeps{
		client:        pbreview.NewReviewCommandServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.CommandMapper(),
		queryMapper:   mapper.QueryMapper(),
		cache:         cache.CommandCache(),
		observability: deps.Observability,
	})
}
