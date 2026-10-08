package order_test

import (
	"context"
	"testing"
	"time"

	order_cache "github.com/MamangRust/microservice-ecommerce-grpc-order/cache"
	order_handler "github.com/MamangRust/microservice-ecommerce-grpc-order/handler"
	order_repo "github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	order_service "github.com/MamangRust/microservice-ecommerce-grpc-order/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderGapiTestSuite struct {
	tests.BaseTestSuite
	queryClient   pb_order.OrderQueryServiceClient
	commandClient pb_order.OrderCommandServiceClient
}

func (s *OrderGapiTestSuite) SetupSuite() {
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
	repos := order_repo.NewRepositories(&order_repo.Deps{
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
		Guards: order_repo.GuardOptions{
			User:        []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Product:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:    []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem:   []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:    []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
			Transaction: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("transaction", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})
	svc := order_service.NewService(&order_service.Deps{
		Cache:         mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := order_handler.NewHandler(&order_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pb_order.RegisterOrderQueryServiceServer(server, handler.OrderQuery)
	pb_order.RegisterOrderCommandServiceServer(server, handler.OrderCommand)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.queryClient = pb_order.NewOrderQueryServiceClient(conn)
	s.commandClient = pb_order.NewOrderCommandServiceClient(conn)
}

func (s *OrderGapiTestSuite) TestOrderGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)

	// 2. Create
	createRes, err := s.commandClient.Create(ctx, &pb_order.CreateOrderRequest{
		UserId:     int32(userID),
		MerchantId: int32(merchID),
		TotalPrice: 10000, Items: []*pb_order.CreateOrderItemRequest{
			{
				ProductId: int32(prodID),
				Quantity:  1,
				// The service must use the server-side product price.
				Price: 1,
			},
		},
		Shipping: &pb_shipping_address.CreateShippingAddressRequest{
			Alamat:         "Test Address",
			Provinsi:       "Test Province",
			Kota:           "Test City",
			Negara:         "Test Country",
			Courier:        "Test Courier",
			ShippingMethod: "Test Method",
			ShippingCost:   1000,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	orderID := createRes.Data.Id

	// 3. FindById
	getRes, err := s.queryClient.FindById(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(userID), getRes.Data.UserId)
	s.Equal(int32(11000), getRes.Data.TotalPrice)

	productRes, err := pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(99), productRes.Data.CountInStock)

	itemClient := pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"])
	itemsRes, err := itemClient.FindOrderItemByOrder(ctx, &pb_order_item.FindByIdOrderItemRequest{Id: orderID})
	s.Require().NoError(err)
	s.Require().Len(itemsRes.Data, 1)
	s.Equal(int32(10000), itemsRes.Data[0].Price)

	// 4. FindAll
	allRes, err := s.queryClient.FindAll(ctx, &pb_order.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := s.queryClient.FindByActive(ctx, &pb_order.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	// Fetch terms first
	s.Require().NotEmpty(itemsRes.Data)
	orderItemID := itemsRes.Data[0].Id

	_, err = s.commandClient.Update(ctx, &pb_order.UpdateOrderRequest{
		OrderId:    orderID,
		UserId:     int32(userID),
		TotalPrice: 15000,
		Items: []*pb_order.UpdateOrderItemRequest{
			{
				OrderItemId: orderItemID,
				ProductId:   int32(prodID),
				Quantity:    2,
				// The service must continue using the server-side product price.
				Price: 1,
			},
		},
		Shipping: &pb_shipping_address.UpdateShippingAddressRequest{
			Alamat:         "Updated Address",
			Provinsi:       "Updated Province",
			Kota:           "Updated City",
			Negara:         "Updated Country",
			Courier:        "Updated Courier",
			ShippingMethod: "Updated Method",
			ShippingCost:   1500,
		},
	})
	s.Require().NoError(err)

	updatedRes, err := s.queryClient.FindById(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(21500), updatedRes.Data.TotalPrice)
	productRes, err = pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(98), productRes.Data.CountInStock)
	itemsRes, err = itemClient.FindOrderItemByOrder(ctx, &pb_order_item.FindByIdOrderItemRequest{Id: orderID})
	s.Require().NoError(err)
	s.Require().Len(itemsRes.Data, 1)
	s.Equal(int32(10000), itemsRes.Data[0].Price)

	// 7. Update items without shipping: persisted shipping and its cost must remain unchanged.
	_, err = s.commandClient.Update(ctx, &pb_order.UpdateOrderRequest{
		OrderId:    orderID,
		UserId:     int32(userID),
		TotalPrice: 1,
		Items: []*pb_order.UpdateOrderItemRequest{
			{OrderItemId: orderItemID, ProductId: int32(prodID), Quantity: 2, Price: 1},
		},
	})
	s.Require().NoError(err)
	unchangedRes, err := s.queryClient.FindById(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(21500), unchangedRes.Data.TotalPrice)

	// 8. Trash
	_, err = s.commandClient.TrashedOrder(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	productRes, err = pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(100), productRes.Data.CountInStock)

	// Repeating trash is rejected and must not change stock again.
	_, err = s.commandClient.TrashedOrder(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().Error(err)
	productRes, err = pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(100), productRes.Data.CountInStock)

	// 9. FindByTrashed
	trashedRes, err := s.queryClient.FindByTrashed(ctx, &pb_order.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 10. Restore
	_, err = s.commandClient.RestoreOrder(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	productRes, err = pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(98), productRes.Data.CountInStock)

	// Repeating restore is rejected and must not reserve stock again.
	_, err = s.commandClient.RestoreOrder(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().Error(err)
	productRes, err = pb_product.NewProductQueryServiceClient(s.Conns["product"]).FindById(ctx, &pb_product.FindByIdProductRequest{Id: int32(prodID)})
	s.Require().NoError(err)
	s.Equal(int32(98), productRes.Data.CountInStock)

	// 11. DeletePermanent
	_, _ = s.commandClient.TrashedOrder(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	_, err = s.commandClient.DeleteOrderPermanent(ctx, &pb_order.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 12. RestoreAll
	_, err = s.commandClient.RestoreAllOrder(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 13. DeleteAll
	_, err = s.commandClient.DeleteAllOrderPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestOrderGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderGapiTestSuite))
}
