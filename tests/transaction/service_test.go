package transaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	trans_cache "github.com/MamangRust/microservice-ecommerce-grpc-transaction/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type TransactionServiceTestSuite struct {
	tests.BaseTestSuite
	svc *service.Service
}

func (s *TransactionServiceTestSuite) SetupSuite() {
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

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Transaction dependencies
	mencache := trans_cache.NewMencache(cacheStore)
	repos := repository.NewRepositories(&repository.Deps{
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

	s.svc = service.NewService(&service.Deps{
		Kafka:         nil,
		Pool:          s.SQLxDB(), // exercises the transactional outbox path (single commit)
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
}

func (s *TransactionServiceTestSuite) TestTransactionLifecycle() {
	ctx := context.Background()

	// 1. Setup Dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()

	userID := s.SeedUser(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	categoryID := s.SeedCategory(ctx)
	productID := s.SeedProduct(ctx, merchantID, categoryID)
	orderID := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, orderID, productID)
	s.SeedShippingAddress(ctx, orderID)

	// 2. Create Transaction
	req := &requests.CreateTransactionRequest{
		UserID:        userID,
		MerchantID:    merchantID,
		OrderID:       orderID,
		Amount:        1000000, // Sufficient amount
		PaymentMethod: "Transfer Bank",
	}
	created, err := s.svc.TransactionCommand.Create(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	transactionID := int(created.TransactionID)

	// 3. FindByID
	found, err := s.svc.TransactionQuery.FindByID(ctx, transactionID)
	s.Require().NoError(err)
	s.Equal("success", found.PaymentStatus)

	// 4. Update
	newPaymentMethod := "GOPAY"
	updateReq := &requests.UpdateTransactionRequest{
		TransactionID: &transactionID,
		MerchantID:    merchantID,
		OrderID:       orderID,
		Amount:        req.Amount,
		PaymentMethod: newPaymentMethod,
	}
	updated, err := s.svc.TransactionCommand.Update(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(newPaymentMethod, updated.PaymentMethod)

	// 5. FindAll
	_, total, err := s.svc.TransactionQuery.FindAll(ctx, &requests.FindAllTransaction{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 6. Trash
	_, err = s.svc.TransactionCommand.Trash(ctx, transactionID)
	s.Require().NoError(err)

	// 7. FindTrashed
	_, totalTrashed, err := s.svc.TransactionQuery.FindTrashed(ctx, &requests.FindAllTransaction{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 8. FindActive
	active, _, err := s.svc.TransactionQuery.FindActive(ctx, &requests.FindAllTransaction{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, tx := range active {
		s.NotEqual(transactionID, int(tx.TransactionID))
	}

	// 9. Restore
	_, err = s.svc.TransactionCommand.Restore(ctx, transactionID)
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, err = s.svc.TransactionCommand.Trash(ctx, transactionID)
	s.Require().NoError(err)
	success, err := s.svc.TransactionCommand.DeletePermanent(ctx, transactionID)
	s.Require().NoError(err)
	s.True(success)

	// Test DeleteByOrderIDPermanent
	oExp := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, oExp, productID)
	s.SeedShippingAddress(ctx, oExp)
	tExp, _ := s.svc.TransactionCommand.Create(ctx, &requests.CreateTransactionRequest{UserID: userID, MerchantID: merchantID, OrderID: oExp, Amount: 1000000, PaymentMethod: "PEXP"})
	s.svc.TransactionCommand.Trash(ctx, int(tExp.TransactionID))
	successDelByOrder, err := s.svc.TransactionCommand.DeleteByOrderIDPermanent(ctx, oExp)
	s.Require().NoError(err)
	s.True(successDelByOrder)

	// 11. RestoreAll & DeleteAll
	o1 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o1, productID)
	s.SeedShippingAddress(ctx, o1)
	t1, _ := s.svc.TransactionCommand.Create(ctx, &requests.CreateTransactionRequest{UserID: userID, MerchantID: merchantID, OrderID: o1, Amount: 1000000, PaymentMethod: "P1"})

	o2 := s.SeedOrder(ctx, userID, merchantID, productID)
	s.SeedOrderItem(ctx, o2, productID)
	s.SeedShippingAddress(ctx, o2)
	t2, _ := s.svc.TransactionCommand.Create(ctx, &requests.CreateTransactionRequest{UserID: userID, MerchantID: merchantID, OrderID: o2, Amount: 1000000, PaymentMethod: "P2"})

	s.svc.TransactionCommand.Trash(ctx, int(t1.TransactionID))
	s.svc.TransactionCommand.Trash(ctx, int(t2.TransactionID))

	resRestoreAll, err := s.svc.TransactionCommand.RestoreAll(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.TransactionCommand.Trash(ctx, int(t1.TransactionID))
	s.svc.TransactionCommand.Trash(ctx, int(t2.TransactionID))

	resDeleteAll, err := s.svc.TransactionCommand.DeleteAll(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestTransactionServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionServiceTestSuite))
}
