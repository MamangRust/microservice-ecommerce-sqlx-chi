package product_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	prod_cache "github.com/MamangRust/microservice-ecommerce-grpc-product/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-product/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-product/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *ProductServiceTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Product dependencies
	mencache := prod_cache.NewMencache(cacheStore)
	guardCategory := resilience.NewDependencyGuard("category", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	repos := repository.NewRepositories(
		s.SQLxDB(),
		pb_category.NewCategoryQueryServiceClient(s.Conns["category"]),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		repository.GuardOptions{
			Category: []adapter.GuardOption{adapter.WithDependencyGuard(guardCategory)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)

	s.svc = service.NewService(&service.Deps{
		Cache:         mencache,
		Repository:    repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *ProductServiceTestSuite) TestProductLifecycle() {
	ctx := context.Background()

	// 1. Setup Dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()

	userID := s.SeedUser(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	categoryID := s.SeedCategory(ctx)

	rating := 5
	slug := "initial-product"

	// 2. Create Product
	req := &requests.CreateProductRequest{
		Name:         "Initial Product",
		Description:  "Product description",
		Price:        10000,
		CountInStock: 100,
		Brand:        "Test Brand",
		Weight:       1000,
		CategoryID:   categoryID,
		MerchantID:   merchantID,
		Rating:       &rating,
		SlugProduct:  &slug,
		ImageProduct: "initial.jpg",
	}
	created, err := s.svc.ProductCommand.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	productID := int(created.ProductID)

	// 3. FindByID
	found, err := s.svc.ProductQuery.FindByID(ctx, productID)
	s.Require().NoError(err)
	s.Equal(req.Name, found.Name)

	// 4. Update
	newName := "Updated Product Name"
	newSlug := "updated-product"
	updateReq := &requests.UpdateProductRequest{
		ProductID:    &productID,
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         newName,
		Description:  req.Description,
		Price:        req.Price,
		CountInStock: 90,
		Brand:        "Updated Brand",
		Weight:       req.Weight,
		Rating:       req.Rating,
		SlugProduct:  &newSlug,
		ImageProduct: "updated.jpg",
	}
	updated, err := s.svc.ProductCommand.Update(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(newName, updated.Name)

	// 5. FindAll
	_, total, err := s.svc.ProductQuery.FindAll(ctx, &requests.FindAllProduct{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 6. Trash
	_, err = s.svc.ProductCommand.Trash(ctx, productID)
	s.Require().NoError(err)

	// 7. FindTrashed
	_, totalTrashed, err := s.svc.ProductQuery.FindTrashed(ctx, &requests.FindAllProduct{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 8. FindActive
	active, _, err := s.svc.ProductQuery.FindActive(ctx, &requests.FindAllProduct{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, p := range active {
		s.NotEqual(productID, int(p.ProductID))
	}

	// 9. Restore
	_, err = s.svc.ProductCommand.Restore(ctx, productID)
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, err = s.svc.ProductCommand.Trash(ctx, productID)
	s.Require().NoError(err)
	success, err := s.svc.ProductCommand.DeletePermanent(ctx, productID)
	s.Require().NoError(err)
	s.True(success)

	// 11. RestoreAll & DeleteAll
	p1, _ := s.svc.ProductCommand.Create(ctx, &requests.CreateProductRequest{MerchantID: merchantID, CategoryID: categoryID, Name: "P1", Description: "D1", Price: 1, CountInStock: 1, Brand: "B1", Weight: 1, Rating: &rating, SlugProduct: &slug, ImageProduct: "I1"})
	p2, _ := s.svc.ProductCommand.Create(ctx, &requests.CreateProductRequest{MerchantID: merchantID, CategoryID: categoryID, Name: "P2", Description: "D2", Price: 2, CountInStock: 2, Brand: "B2", Weight: 2, Rating: &rating, SlugProduct: &slug, ImageProduct: "I2"})

	s.svc.ProductCommand.Trash(ctx, int(p1.ProductID))
	s.svc.ProductCommand.Trash(ctx, int(p2.ProductID))

	resRestoreAll, err := s.svc.ProductCommand.RestoreAll(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.ProductCommand.Trash(ctx, int(p1.ProductID))
	s.svc.ProductCommand.Trash(ctx, int(p2.ProductID))

	resDeleteAll, err := s.svc.ProductCommand.DeleteAll(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestProductServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductServiceTestSuite))
}
