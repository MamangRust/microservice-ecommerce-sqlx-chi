package order_test

import (
	"context"
	"testing"
	"time"

	order_cache "github.com/MamangRust/microservice-ecommerce-grpc-order/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type OrderServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *OrderServiceTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupTransactionService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Order dependencies
	mencache := order_cache.NewMencache(cacheStore)
	repos := repository.NewRepositories(&repository.Deps{
		DB:                       s.SQLxDB(),
		UserQueryClient:          pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		ProductQueryClient:       pb_product.NewProductQueryServiceClient(s.Conns["product"]),
		ProductCommandClient:     pb_product.NewProductCommandServiceClient(s.Conns["product"]),
		MerchantQueryClient:      pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		OrderItemQueryClient:     pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		OrderItemCommandClient:   pb_order_item.NewOrderItemCommandServiceClient(s.Conns["order-item"]),
		ShippingCommandClient:    pb_shipping_address.NewShippingCommandServiceClient(s.Conns["shipping-address"]),
		ShippingQueryClient:      pb_shipping_address.NewShippingQueryServiceClient(s.Conns["shipping-address"]),
		TransactionCommandClient: pb_transaction.NewTransactionCommandServiceClient(s.Conns["transaction"]),
		Guards: repository.GuardOptions{
			User:        []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Product:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:    []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem:   []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:    []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
			Transaction: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("transaction", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})

	s.svc = service.NewService(&service.Deps{
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *OrderServiceTestSuite) TestOrderLifecycle() {
	ctx := context.Background()

	// 1. Setup Dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()

	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)

	// 2. Create Order
	req := &requests.CreateOrderRequest{
		UserID:     userID,
		MerchantID: merchID,
		TotalPrice: 20000,
		Items: []requests.CreateOrderItemRequest{
			{
				ProductID: prodID,
				Quantity:  1,
				Price:     10000,
			},
		},
		ShippingAddress: requests.CreateShippingAddressRequest{
			Alamat:         "Test Address",
			Provinsi:       "West Java",
			Kota:           "Bandung",
			Courier:        "JNE",
			ShippingMethod: "REG",
			ShippingCost:   10000,
			Negara:         "Indonesia",
		},
	}

	created, err := s.svc.OrderCommand.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	orderID := int(created.OrderID)

	// 3. FindByID
	found, err := s.svc.OrderQuery.FindByID(ctx, orderID)
	s.Require().NoError(err)
	s.Equal(int32(req.TotalPrice), found.TotalPrice)

	// 4. FindAll
	_, total, err := s.svc.OrderQuery.FindAll(ctx, &requests.FindAllOrder{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 5. Trash
	_, err = s.svc.OrderCommand.Trash(ctx, orderID)
	s.Require().NoError(err)

	// 6. FindTrashed
	_, totalTrashed, err := s.svc.OrderQuery.FindTrashed(ctx, &requests.FindAllOrder{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 7. FindActive
	active, _, err := s.svc.OrderQuery.FindActive(ctx, &requests.FindAllOrder{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, o := range active {
		s.NotEqual(orderID, int(o.OrderID))
	}

	// 8. Restore
	_, err = s.svc.OrderCommand.Restore(ctx, orderID)
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, err = s.svc.OrderCommand.Trash(ctx, orderID)
	s.Require().NoError(err)
	success, err := s.svc.OrderCommand.DeletePermanent(ctx, orderID)
	s.Require().NoError(err)
	s.True(success)

	// 10. RestoreAll & DeleteAll
	o1, _ := s.svc.OrderCommand.Create(ctx, req)
	o2, _ := s.svc.OrderCommand.Create(ctx, req)

	s.svc.OrderCommand.Trash(ctx, int(o1.OrderID))
	s.svc.OrderCommand.Trash(ctx, int(o2.OrderID))

	resRestoreAll, err := s.svc.OrderCommand.RestoreAll(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.OrderCommand.Trash(ctx, int(o1.OrderID))
	s.svc.OrderCommand.Trash(ctx, int(o2.OrderID))

	// Note: We use the correct method name from the transaction dependency
	// if applicable, but this is the Order service test.
	// Ensuring Cleanup.
	resDeleteAll, err := s.svc.OrderCommand.DeleteAll(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestOrderServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderServiceTestSuite))
}
