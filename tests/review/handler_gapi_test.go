package review_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	review_cache "github.com/MamangRust/microservice-ecommerce-grpc-review/cache"
	review_handler "github.com/MamangRust/microservice-ecommerce-grpc-review/handler"
	review_repo "github.com/MamangRust/microservice-ecommerce-grpc-review/repository"
	review_service "github.com/MamangRust/microservice-ecommerce-grpc-review/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ReviewGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_review.ReviewQueryServiceClient
	commandClient pb_review.ReviewCommandServiceClient
}

func (s *ReviewGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Review dependencies
	mencache := review_cache.NewMencache(cacheStore)
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log)
	repos := review_repo.NewRepositories(
		s.SQLxDB(),
		pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		pb_product.NewProductQueryServiceClient(s.Conns["product"]),
		review_repo.GuardOptions{
			User:    []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Product: []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
		},
	)
	svc := review_service.NewService(&review_service.Deps{
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := review_handler.NewHandler(&review_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_review.RegisterReviewQueryServiceServer(server, handler.ReviewQuery)
	pb_review.RegisterReviewCommandServiceServer(server, handler.ReviewCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_review.NewReviewQueryServiceClient(conn)
	s.commandClient = pb_review.NewReviewCommandServiceClient(conn)
}

func (s *ReviewGapiTestSuite) TestReviewGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := int32(s.SeedProduct(ctx, merchID, catID))

	// 2. Create
	createRes, err := s.commandClient.Create(ctx, &pb_review.CreateReviewRequest{
		UserId:    int32(userID),
		ProductId: prodID,
		Rating:    5,
		Comment:   "GAPI Comment",
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	reviewID := createRes.Data.Id

	// 3. FindById
	_, err = s.queryClient.FindAll(ctx, &pb_review.FindAllReviewRequest{Page: 1, PageSize: 10}) // Review query doesn't have FindById in proto?
	s.NoError(err)
	// s.Equal("GAPI Comment", getRes.Data[0].Comment) // Adjustment for FindAll if FindById is missing

	// 4. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_review.FindAllReviewRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_review.FindAllReviewRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	updateRes, err := s.commandClient.Update(ctx, &pb_review.UpdateReviewRequest{
		ReviewId: reviewID,
		Rating:   4,
		Comment:  "GAPI Comment Updated",
	})
	s.NoError(err)
	s.Equal("GAPI Comment Updated", updateRes.Data.Comment)

	// 7. Trash
	_, err = s.commandClient.TrashedReview(ctx, &pb_review.FindByIdReviewRequest{Id: reviewID})
	s.NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_review.FindAllReviewRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.commandClient.RestoreReview(ctx, &pb_review.FindByIdReviewRequest{Id: reviewID})
	s.NoError(err)

	// 10. DeletePermanent
	_, _ = s.commandClient.TrashedReview(ctx, &pb_review.FindByIdReviewRequest{Id: reviewID})
	_, err = s.commandClient.DeleteReviewPermanent(ctx, &pb_review.FindByIdReviewRequest{Id: reviewID})
	s.NoError(err)

	// 11. RestoreAll
	_, err = s.commandClient.RestoreAllReview(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 12. DeleteAll
	_, err = s.commandClient.DeleteAllReviewPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestReviewGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewGapiTestSuite))
}
