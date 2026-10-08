package transaction_test

import (
	"context"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TransactionRepositoryTestSuite struct {
	tests.BaseTestSuite
	repo *repository.Repositories
}

func (s *TransactionRepositoryTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupShippingAddressService()
	s.SetupOrderItemService()
	s.SetupOrderService()

	s.repo = repository.NewRepositories(&repository.Deps{
		DB:                   s.SQLxDB(),
		UserQueryClient:      pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		MerchantQueryClient:  pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		OrderQueryClient:     pb_order.NewOrderQueryServiceClient(s.Conns["order"]),
		OrderItemQueryClient: pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		ShippingQueryClient:  pb_shipping_address.NewShippingQueryServiceClient(s.Conns["shipping-address"]),
		Guards: repository.GuardOptions{
			User:      []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			Order:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})
}

func (s *TransactionRepositoryTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *TransactionRepositoryTestSuite) TestTransactionLifecycle() {
	ctx := context.Background()

	// Seed dependencies
	userID := s.SeedUser(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, merchID, catID)
	orderID := s.SeedOrder(ctx, userID, merchID, prodID)

	pageReq := &requests.FindAllTransaction{Search: "", Page: 1, PageSize: 10}

	// 1. Create Transaction
	status := "paid"
	req := &requests.CreateTransactionRequest{
		UserID:        userID,
		OrderID:       orderID,
		MerchantID:    merchID,
		PaymentMethod: "Credit Card",
		PaymentStatus: &status,
		Amount:        100000,
	}
	created, err := s.repo.TransactionCommand.Create(ctx, req)
	s.NoError(err)
	s.NotNil(created)
	s.Equal(int32(orderID), created.OrderID)
	txnID := int(created.TransactionID)

	// 2. FindByID
	found, err := s.repo.TransactionQuery.FindByID(ctx, txnID)
	s.NoError(err)
	s.Equal(created.Amount, found.Amount)

	// 3. FindByOrderID
	foundByOrder, err := s.repo.TransactionQuery.FindByOrderID(ctx, orderID)
	s.NoError(err)
	s.NotNil(foundByOrder)

	// 4. FindAll
	all, err := s.repo.TransactionQuery.FindAll(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(all)

	// 5. FindActive (before trash)
	active, err := s.repo.TransactionQuery.FindActive(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(active)

	// 6. Update
	updatedStatus := "settled"
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID: &txnID,
		OrderID:       orderID,
		MerchantID:    merchID,
		PaymentMethod: "Debit Card",
		Amount:        120000,
		PaymentStatus: &updatedStatus,
	}
	updated, err := s.repo.TransactionCommand.Update(ctx, updateReq)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(120000), updated.Amount)

	// 7. Trash
	trashed, err := s.repo.TransactionCommand.Trash(ctx, txnID)
	s.NoError(err)
	s.NotNil(trashed)

	// 8. FindTrashed
	trashedItems, err := s.repo.TransactionQuery.FindTrashed(ctx, pageReq)
	s.NoError(err)
	s.NotEmpty(trashedItems)

	// 9. FindActive after trash — should NOT include
	activeAfterTrash, err := s.repo.TransactionQuery.FindActive(ctx, pageReq)
	s.NoError(err)
	for _, item := range activeAfterTrash {
		s.NotEqual(txnID, int(item.TransactionID))
	}

	// 10. Restore
	restored, err := s.repo.TransactionCommand.Restore(ctx, txnID)
	s.NoError(err)
	s.NotNil(restored)

	// 11. Trash again then DeletePermanent
	_, err = s.repo.TransactionCommand.Trash(ctx, txnID)
	s.Require().NoError(err)

	success, err := s.repo.TransactionCommand.DeletePermanent(ctx, txnID)
	s.NoError(err)
	s.True(success)

	// 12. RestoreAll
	status2 := "pending"
	second, _ := s.repo.TransactionCommand.Create(ctx, &requests.CreateTransactionRequest{
		UserID: userID, OrderID: orderID, MerchantID: merchID,
		PaymentMethod: "Bank Transfer", PaymentStatus: &status2, Amount: 50000,
	})
	s.repo.TransactionCommand.Trash(ctx, int(second.TransactionID))

	resRestore, err := s.repo.TransactionCommand.RestoreAll(ctx)
	s.NoError(err)
	s.True(resRestore)

	// 13. DeleteAll
	s.repo.TransactionCommand.Trash(ctx, int(second.TransactionID))
	resDelete, err := s.repo.TransactionCommand.DeleteAll(ctx)
	s.NoError(err)
	s.True(resDelete)
}

func TestTransactionRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionRepositoryTestSuite))
}
