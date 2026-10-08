package reviewdetailhandler

import (
	reviewdetail_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/review_detail"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	reviewapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/review_detail"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	pbreviewdetail "github.com/MamangRust/microservice-ecommerce-grpc-pb/review_detail"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type DepsReviewDetail struct {
	Client        *grpc.ClientConn
	Router        chi.Router
	Logger        logger.LoggerInterface
	Cache         *cache.CacheStore
	Upload        upload_image.ImageUploads
	Observability observability.TraceLoggerObservability
}

func RegisterReviewDetailHandler(deps *DepsReviewDetail) {
	mapper := apimapper.NewReviewDetailResponseMapper()
	reviewMapper := reviewapimapper.NewReviewResponseMapper()
	cache := reviewdetail_cache.NewReviewDetailMencache(deps.Cache)

	NewReviewDetailQueryHandleApi(&reviewDetailQueryHandleDeps{
		client:        pbreviewdetail.NewReviewDetailQueryServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.QueryMapper(),
		cache:         cache.QueryCache(),
		observability: deps.Observability,
	})

	NewReviewDetailCommandHandleApi(&reviewDetailCommandHandleDeps{
		client:        pbreviewdetail.NewReviewDetailCommandServiceClient(deps.Client),
		router:        deps.Router,
		logger:        deps.Logger,
		mapper:        mapper.CommandMapper(),
		queryMapper:   mapper.QueryMapper(),
		reviewMapper:  reviewMapper.CommandMapper(),
		cache:         cache.CommandCache(),
		upload:        deps.Upload,
		observability: deps.Observability,
	})
}
