package order_test

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type OrderRepositoryTestSuite struct {
	tests.BaseTestSuite
	repo *repository.Repositories
}

func (s *OrderRepositoryTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupTransactionService()

	s.repo = repository.NewRepositories(&repository.Deps{
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
}

func (s *OrderRepositoryTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *OrderRepositoryTestSuite) TestOrderLifecycle() {
	ctx := context.Background()

	// Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	s.SeedProduct(ctx, merchID, catID)

	pageReq := &requests.FindAllOrder{Search: "", Page: 1, PageSize: 10}

	// 1. Create Order
	req := &requests.CreateOrderRecordRequest{
		UserID:     userID,
		MerchantID: merchID,
		TotalPrice: 5000,
	}
	created, err := s.repo.OrderCommand.Create(ctx, req)
	s.NoError(err)
	s.NotNil(created)
	s.Equal(int32(userID), created.UserID)
	orderID := int(created.OrderID)

	// 2. FindByID
	found, err := s.repo.OrderQuery.FindByID(ctx, orderID)
	s.NoError(err)
	s.Equal(int32(5000), found.TotalPrice)

	// 3. FindAll
	all, err := s.repo.OrderQuery.FindAll(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(all)

	// 4. FindActive (before trash)
	active, err := s.repo.OrderQuery.FindActive(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(active)

	// 5. Update
	updateReq := &requests.UpdateOrderRecordRequest{
		OrderID:    orderID,
		MerchantID: merchID,
		UserID:     userID,
		TotalPrice: 10000,
	}
	updated, err := s.repo.OrderCommand.Update(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(10000), updated.TotalPrice)

	// 6. Trash
	trashed, err := s.repo.OrderCommand.Trash(ctx, orderID)
	s.NoError(err)
	s.NotNil(trashed)

	// 7. FindTrashed
	trashedItems, err := s.repo.OrderQuery.FindTrashed(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(trashedItems)

	// 8. FindActive after trash — should NOT include
	activeAfterTrash, err := s.repo.OrderQuery.FindActive(ctx, pageReq)
	s.NoError(err)
	for _, item := range activeAfterTrash {
		s.NotEqual(orderID, int(item.OrderID))
	}

	// 9. Restore
	restored, err := s.repo.OrderCommand.Restore(ctx, orderID)
	s.NoError(err)
	s.NotNil(restored)

	// 10. Trash again then DeletePermanent
	_, err = s.repo.OrderCommand.Trash(ctx, orderID)
	s.Require().NoError(err)

	success, err := s.repo.OrderCommand.DeletePermanent(ctx, orderID)
	s.NoError(err)
	s.True(success)

	// 11. RestoreAll
	second, _ := s.repo.OrderCommand.Create(ctx, &requests.CreateOrderRecordRequest{
		UserID: userID, MerchantID: merchID, TotalPrice: 2000,
	})
	s.repo.OrderCommand.Trash(ctx, int(second.OrderID))

	resRestore, err := s.repo.OrderCommand.RestoreAll(ctx)
	s.NoError(err)
	s.True(resRestore)

	// 12. DeleteAll
	s.repo.OrderCommand.Trash(ctx, int(second.OrderID))
	resDelete, err := s.repo.OrderCommand.DeleteAll(ctx)
	s.NoError(err)
	s.True(resDelete)
}

func TestOrderRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderRepositoryTestSuite))
}
