package apps

import (
	"fmt"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-order/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/service"
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	pb_order "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	pb_order_item "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	pb_product "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	pb_shipping_address "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	pb_transaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-pkg/server"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pkgresilience "github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	guard := pkgresilience.NewDependencyGuardInterceptor(srv.Logger)

	userAddr := viper.GetString("GRPC_USER_ADDR")

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	userQueryClient := pb_user.NewUserQueryServiceClient(userConn)

	productAddr := viper.GetString("GRPC_PRODUCT_ADDR")

	productConn, err := grpc.NewClient(productAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}
	productQueryClient := pb_product.NewProductQueryServiceClient(productConn)
	productCommandClient := pb_product.NewProductCommandServiceClient(productConn)

	merchantAddr := viper.GetString("GRPC_MERCHANT_ADDR")
	if merchantAddr == "" {
		merchantAddr = "merchant:50055"
	}
	merchantConn, err := grpc.NewClient(merchantAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to merchant service: %w", err)
	}
	merchantQueryClient := pb_merchant.NewMerchantQueryServiceClient(merchantConn)

	orderItemAddr := viper.GetString("GRPC_ORDER_ITEM_ADDR")
	if orderItemAddr == "" {
		orderItemAddr = "order-item:50056"
	}
	orderItemConn, err := grpc.NewClient(orderItemAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order_item service: %w", err)
	}
	orderItemQueryClient := pb_order_item.NewOrderItemQueryServiceClient(orderItemConn)
	orderItemCommandClient := pb_order_item.NewOrderItemCommandServiceClient(orderItemConn)

	shippingAddr := viper.GetString("GRPC_SHIPPING_ADDRESS_ADDR")
	if shippingAddr == "" {
		shippingAddr = "shipping_address:50063"
	}
	shippingConn, err := grpc.NewClient(shippingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to shipping_address service: %w", err)
	}
	shippingCommandClient := pb_shipping_address.NewShippingCommandServiceClient(shippingConn)

	transactionAddr := viper.GetString("GRPC_TRANSACTION_ADDR")
	if transactionAddr == "" {
		transactionAddr = "transaction:50061"
	}
	transactionConn, err := grpc.NewClient(transactionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(guard.UnaryInterceptor()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to transaction service: %w", err)
	}
	transactionCommandClient := pb_transaction.NewTransactionCommandServiceClient(transactionConn)

	repos := repository.NewRepositories(&repository.Deps{
		DB:                       srv.DB,
		UserQueryClient:          userQueryClient,
		ProductQueryClient:       productQueryClient,
		ProductCommandClient:     productCommandClient,
		MerchantQueryClient:      merchantQueryClient,
		OrderItemQueryClient:     orderItemQueryClient,
		OrderItemCommandClient:   orderItemCommandClient,
		ShippingCommandClient:    shippingCommandClient,
		ShippingQueryClient:      pb_shipping_address.NewShippingQueryServiceClient(shippingConn),
		TransactionCommandClient: transactionCommandClient,
		Guards: repository.GuardOptions{
			User: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			Product: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			Merchant: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			OrderItem: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("order-item", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			Shipping: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping-address", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			Transaction: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("transaction", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	})

	obs, _ := observability.NewObservability("order-server", srv.Logger)
	cache := cache.NewMencache(srv.CacheStore)
	myKafka := kafka.NewKafka(srv.Logger, []string{viper.GetString("KAFKA_BROKERS")})

	svc := service.NewService(&service.Deps{
		Kafka:         myKafka,
		Cache:         cache,
		Logger:        srv.Logger,
		Repositories:  repos,
		Observability: obs,
	})

	h := handler.NewHandler(&handler.Deps{Service: svc, Logger: srv.Logger})

	// Start the outbox relay so stats events committed with the order flow are
	// published to Kafka with durable retry and dead-letter semantics (F3).
	go svc.Outbox.Start(srv.Ctx, service.OutboxRelayInterval, service.OutboxRelayBatchSize)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb_order.RegisterOrderQueryServiceServer(gs, h.OrderQuery)
		pb_order.RegisterOrderCommandServiceServer(gs, h.OrderCommand)
	}

	return srv, nil
}
