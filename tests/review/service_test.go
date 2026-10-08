package review_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	review_cache "github.com/MamangRust/microservice-ecommerce-grpc-review/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ReviewServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *ReviewServiceTestSuite) SetupSuite() {
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
	repos := repository.NewRepositories(
		s.SQLxDB(),
		pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		pb_product.NewProductQueryServiceClient(s.Conns["product"]),
		repository.GuardOptions{
			User:    []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Product: []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
		},
	)

	s.svc = service.NewService(&service.Deps{
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *ReviewServiceTestSuite) TestReviewLifecycle() {
	ctx := context.Background()

	// 1. Setup Dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()

	userID := s.SeedUser(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	categoryID := s.SeedCategory(ctx)
	productID := s.SeedProduct(ctx, merchantID, categoryID)

	// 2. Create Review
	req := &requests.CreateReviewRequest{
		UserID:    userID,
		ProductID: productID,
		Rating:    5,
		Comment:   "Excellent product!",
	}
	created, err := s.svc.ReviewCommand.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	reviewID := int(created.ReviewID)

	// 3. FindByID
	found, err := s.svc.ReviewQuery.FindByID(ctx, reviewID)
	s.Require().NoError(err)
	s.Equal(req.Comment, found.Comment)

	// 4. Update
	newComment := "Updated comment: Still excellent!"
	updateReq := &requests.UpdateReviewRequest{
		ReviewID: &reviewID,
		Rating:   req.Rating,
		Comment:  newComment,
	}
	updated, err := s.svc.ReviewCommand.Update(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(newComment, updated.Comment)

	// 5. FindAll
	_, total, err := s.svc.ReviewQuery.FindAll(ctx, &requests.FindAllReview{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 6. Trash
	_, err = s.svc.ReviewCommand.Trash(ctx, reviewID)
	s.Require().NoError(err)

	// 7. FindTrashed
	_, totalTrashed, err := s.svc.ReviewQuery.FindTrashed(ctx, &requests.FindAllReview{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 8. FindActive
	active, _, err := s.svc.ReviewQuery.FindActive(ctx, &requests.FindAllReview{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, r := range active {
		s.NotEqual(reviewID, int(r.ReviewID))
	}

	// 9. Restore
	_, err = s.svc.ReviewCommand.Restore(ctx, reviewID)
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, err = s.svc.ReviewCommand.Trash(ctx, reviewID)
	s.Require().NoError(err)
	success, err := s.svc.ReviewCommand.DeletePermanent(ctx, reviewID)
	s.Require().NoError(err)
	s.True(success)

	// 11. RestoreAll & DeleteAll
	r1, _ := s.svc.ReviewCommand.Create(ctx, &requests.CreateReviewRequest{UserID: userID, ProductID: productID, Rating: 4, Comment: "C1"})
	r2, _ := s.svc.ReviewCommand.Create(ctx, &requests.CreateReviewRequest{UserID: userID, ProductID: productID, Rating: 4, Comment: "C2"})

	s.svc.ReviewCommand.Trash(ctx, int(r1.ReviewID))
	s.svc.ReviewCommand.Trash(ctx, int(r2.ReviewID))

	resRestoreAll, err := s.svc.ReviewCommand.RestoreAll(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.ReviewCommand.Trash(ctx, int(r1.ReviewID))
	s.svc.ReviewCommand.Trash(ctx, int(r2.ReviewID))

	resDeleteAll, err := s.svc.ReviewCommand.DeleteAll(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestReviewServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewServiceTestSuite))
}
