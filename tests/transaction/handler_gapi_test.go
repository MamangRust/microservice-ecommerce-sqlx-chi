package transaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	trans_cache "github.com/MamangRust/microservice-ecommerce-grpc-transaction/cache"
	trans_handler "github.com/MamangRust/microservice-ecommerce-grpc-transaction/handler"
	trans_repo "github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	trans_service "github.com/MamangRust/microservice-ecommerce-grpc-transaction/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TransactionGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_transaction.TransactionQueryServiceClient
	commandClient pb_transaction.TransactionCommandServiceClient
}

func (s *TransactionGapiTestSuite) SetupSuite() {
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
	repos := trans_repo.NewRepositories(&trans_repo.Deps{
		DB:                   s.SQLxDB(),
		UserQueryClient:      pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		MerchantQueryClient:  pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		OrderQueryClient:     pb_order.NewOrderQueryServiceClient(s.Conns["order"]),
		OrderItemQueryClient: pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		ShippingQueryClient:  pb_shipping_address.NewShippingQueryServiceClient(s.Conns["shipping-address"]),
		Guards: trans_repo.GuardOptions{
			User:      []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			Order:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})
	svc := trans_service.NewService(&trans_service.Deps{
		Kafka:         nil,
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := trans_handler.NewHandler(&trans_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_transaction.RegisterTransactionQueryServiceServer(server, handler.TransactionQuery)
	pb_transaction.RegisterTransactionCommandServiceServer(server, handler.TransactionCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_transaction.NewTransactionQueryServiceClient(conn)
	s.commandClient = pb_transaction.NewTransactionCommandServiceClient(conn)
}

func (s *TransactionGapiTestSuite) TestTransactionGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, merchID, catID)
	orderID := s.SeedOrder(ctx, userID, merchID, prodID)

	// 2. Create
	createRes, err := s.commandClient.Create(ctx, &pb_transaction.CreateTransactionRequest{
		UserId:        int32(userID),
		MerchantId:    int32(merchID),
		OrderId:       int32(orderID),
		PaymentMethod: "E-Wallet",
		PaymentStatus: "pending",
		Amount:        100000,
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	transID := createRes.Data.Id

	// 3. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_transaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)
	s.Equal("E-Wallet", getRes.Data.PaymentMethod)

	// 4. FindAll
	allRes, err := s.queryClient.FindAllTransactions(ctx, &pb_transaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_transaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	updateRes, err := s.commandClient.Update(ctx, &pb_transaction.UpdateTransactionRequest{
		TransactionId: transID,
		PaymentMethod: "Credit Card",
		PaymentStatus: "success",
		Amount:        150000,
	})
	s.Require().NoError(err)
	s.Equal("Credit Card", updateRes.Data.PaymentMethod)

	// 7. Trash
	_, err = s.commandClient.TrashedTransaction(ctx, &pb_transaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_transaction.FindAllTransactionRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.commandClient.RestoreTransaction(ctx, &pb_transaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, _ = s.commandClient.TrashedTransaction(ctx, &pb_transaction.FindByIdTransactionRequest{Id: transID})
	_, err = s.commandClient.DeleteTransactionPermanent(ctx, &pb_transaction.FindByIdTransactionRequest{Id: transID})
	s.Require().NoError(err)

	// 11. RestoreAll
	_, err = s.commandClient.RestoreAllTransaction(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 12. DeleteAll
	_, err = s.commandClient.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestTransactionGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionGapiTestSuite))
}
