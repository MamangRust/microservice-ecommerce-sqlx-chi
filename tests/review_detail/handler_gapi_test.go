package review_detail_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review_detail"
	detail_cache "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/cache"
	detail_handler "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/handler"
	detail_repo "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/repository"
	detail_service "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/service"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type ReviewDetailGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_review_detail.ReviewDetailQueryServiceClient
	commandClient pb_review_detail.ReviewDetailCommandServiceClient
}

func (s *ReviewDetailGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupOrderService()
	s.SetupReviewService()
	s.SetupReviewDetailService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Detail dependencies
	mencache := detail_cache.NewMencache(cacheStore)
	repos := detail_repo.NewRepositories(s.SQLxDB())
	svc := detail_service.NewService(&detail_service.Deps{
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := detail_handler.NewHandler(&detail_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_review_detail.RegisterReviewDetailQueryServiceServer(server, handler.ReviewDetailQuery)
	pb_review_detail.RegisterReviewDetailCommandServiceServer(server, handler.ReviewDetailCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_review_detail.NewReviewDetailQueryServiceClient(conn)
	s.commandClient = pb_review_detail.NewReviewDetailCommandServiceClient(conn)
}

func (s *ReviewDetailGapiTestSuite) TestGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	categoryID := s.SeedCategory(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	productID := s.SeedProduct(ctx, merchantID, categoryID)
	reviewID := s.SeedReview(ctx, userID, productID)

	// Create Detail
	createRes, err := s.commandClient.Create(ctx, &pb_review_detail.CreateReviewDetailRequest{
		ReviewId: int32(reviewID),
		Type:     "photo",
		Url:      "http://example.com/image.jpg",
		Caption:  "GAPI Detail Comment",
	})
	s.NoError(err)
	s.NotNil(createRes)

	detailID := createRes.Data.Id

	// Get
	getRes, err := s.queryClient.FindById(ctx, &pb_review_detail.FindByIdReviewDetailRequest{Id: detailID})
	s.NoError(err)
	s.Equal("GAPI Detail Comment", getRes.Data.Caption)
}

func TestReviewDetailGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewDetailGapiTestSuite))
}
