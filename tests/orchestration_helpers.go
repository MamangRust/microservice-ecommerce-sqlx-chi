package tests

import (
	"bytes"
	"mime/multipart"
	"time"

	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/auth"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_award"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"

	// Per-service generated schemas (test database holds every table)
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Role
	role_cache "github.com/MamangRust/microservice-ecommerce-grpc-role/cache"
	role_handler "github.com/MamangRust/microservice-ecommerce-grpc-role/handler"
	role_repo "github.com/MamangRust/microservice-ecommerce-grpc-role/repository"
	role_service "github.com/MamangRust/microservice-ecommerce-grpc-role/service"

	// User
	user_cache "github.com/MamangRust/microservice-ecommerce-grpc-user/cache"
	user_handler "github.com/MamangRust/microservice-ecommerce-grpc-user/handler"
	user_repo "github.com/MamangRust/microservice-ecommerce-grpc-user/repository"
	user_service "github.com/MamangRust/microservice-ecommerce-grpc-user/service"

	// Auth
	auth_cache "github.com/MamangRust/microservice-ecommerce-auth/cache"
	auth_handler "github.com/MamangRust/microservice-ecommerce-auth/handler"
	auth_repo "github.com/MamangRust/microservice-ecommerce-auth/repository"
	auth_service "github.com/MamangRust/microservice-ecommerce-auth/service"

	// Banner
	banner_cache "github.com/MamangRust/microservice-ecommerce-grpc-banner/cache"
	banner_handler "github.com/MamangRust/microservice-ecommerce-grpc-banner/handler"
	banner_repo "github.com/MamangRust/microservice-ecommerce-grpc-banner/repository"
	banner_service "github.com/MamangRust/microservice-ecommerce-grpc-banner/service"

	// Slider
	slider_cache "github.com/MamangRust/microservice-ecommerce-grpc-slider/cache"
	slider_handler "github.com/MamangRust/microservice-ecommerce-grpc-slider/handler"
	slider_repo "github.com/MamangRust/microservice-ecommerce-grpc-slider/repository"
	slider_service "github.com/MamangRust/microservice-ecommerce-grpc-slider/service"

	// Category
	category_cache "github.com/MamangRust/microservice-ecommerce-grpc-category/cache"
	category_handler "github.com/MamangRust/microservice-ecommerce-grpc-category/handler"
	category_repo "github.com/MamangRust/microservice-ecommerce-grpc-category/repository"
	category_service "github.com/MamangRust/microservice-ecommerce-grpc-category/service"

	// Product
	product_cache "github.com/MamangRust/microservice-ecommerce-grpc-product/cache"
	product_handler "github.com/MamangRust/microservice-ecommerce-grpc-product/handler"
	product_repo "github.com/MamangRust/microservice-ecommerce-grpc-product/repository"
	product_service "github.com/MamangRust/microservice-ecommerce-grpc-product/service"

	// Cart
	cart_cache "github.com/MamangRust/microservice-ecommerce-grpc-cart/cache"
	cart_handler "github.com/MamangRust/microservice-ecommerce-grpc-cart/handler"
	cart_repo "github.com/MamangRust/microservice-ecommerce-grpc-cart/repository"
	cart_service "github.com/MamangRust/microservice-ecommerce-grpc-cart/service"

	// Merchant
	merchant_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant/cache"
	merchant_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant/handler"
	merchant_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant/repository"
	merchant_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant/service"

	// Order
	order_cache "github.com/MamangRust/microservice-ecommerce-grpc-order/cache"
	order_handler "github.com/MamangRust/microservice-ecommerce-grpc-order/handler"
	order_repo "github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	order_service "github.com/MamangRust/microservice-ecommerce-grpc-order/service"

	// Merchant Award
	merchant_award_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/cache"
	merchant_award_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/handler"
	merchant_award_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/repository"
	merchant_award_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/service"

	// Merchant Business
	merchant_business_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/cache"
	merchant_business_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/handler"
	merchant_business_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/repository"
	merchant_business_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/service"

	// Transaction
	transaction_cache "github.com/MamangRust/microservice-ecommerce-grpc-transaction/cache"
	transaction_handler "github.com/MamangRust/microservice-ecommerce-grpc-transaction/handler"
	transaction_repo "github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	transaction_service "github.com/MamangRust/microservice-ecommerce-grpc-transaction/service"

	// Merchant Detail
	merchant_detail_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/cache"
	merchant_detail_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/handler"
	merchant_detail_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/repository"
	merchant_detail_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/service"

	// Merchant Policy
	merchant_policy_cache "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/cache"
	merchant_policy_handler "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/handler"
	merchant_policy_repo "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/repository"
	merchant_policy_service "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/service"

	// Shipping Address
	shipping_address_cache "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/cache"
	shipping_address_handler "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/handler"
	shipping_address_repo "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/repository"
	shipping_address_service "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/service"

	// Order Item
	order_item_cache "github.com/MamangRust/microservice-ecommerce-grpc-order-item/cache"
	order_item_handler "github.com/MamangRust/microservice-ecommerce-grpc-order-item/handler"
	order_item_repo "github.com/MamangRust/microservice-ecommerce-grpc-order-item/repository"
	order_item_service "github.com/MamangRust/microservice-ecommerce-grpc-order-item/service"

	// Review
	review_cache "github.com/MamangRust/microservice-ecommerce-grpc-review/cache"
	review_handler "github.com/MamangRust/microservice-ecommerce-grpc-review/handler"
	review_repo "github.com/MamangRust/microservice-ecommerce-grpc-review/repository"
	review_service "github.com/MamangRust/microservice-ecommerce-grpc-review/service"

	// Review Detail
	review_detail_cache "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/cache"
	review_detail_handler "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/handler"
	review_detail_repo "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/repository"
	review_detail_service "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/service"
)

func (s *BaseTestSuite) SetupRoleService() {
	cacheStore := s.GetCacheStore()
	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(s.ts.SQLxDB())
	roleSvc := role_service.NewService(&role_service.Deps{
		Repository:    roleRepos,
		Logger:        s.Log,
		Cache:         roleMencache,
		Observability: s.Obs,
	})
	roleGapi := role_handler.NewHandler(&role_handler.Deps{
		Service: roleSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_role.RegisterRoleQueryServiceServer(server, roleGapi.RoleQuery)
	pb_role.RegisterRoleCommandServiceServer(server, roleGapi.RoleCommand)
	pb_user_role.RegisterUserRoleServiceServer(server, roleGapi.UserRole)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["role"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupUserService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}

	cacheStore := s.GetCacheStore()
	hasher := hash.NewHashingPassword()

	userMencache := user_cache.NewMencache(cacheStore)
	roleQueryClient := pb_role.NewRoleQueryServiceClient(s.Conns["role"])
	userRoleClient := pb_user_role.NewUserRoleServiceClient(s.Conns["role"])
	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, s.Log)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, s.Log)
	userRepos := user_repo.NewRepositories(&user_repo.Deps{
		Db:       s.ts.SQLxDB(),
		Role:     roleQueryClient,
		UserRole: userRoleClient,
		Guards: user_repo.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	})
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        s.Log,
		Hash:          hasher,
		Cache:         userMencache,
		Observability: s.Obs,
	})
	userGapi := user_handler.NewHandler(&user_handler.Deps{
		Service: userSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_user.RegisterUserQueryServiceServer(server, userGapi.UserQuery)
	pb_user.RegisterUserCommandServiceServer(server, userGapi.UserCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["user"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupAuthService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}

	cacheStore := s.GetCacheStore()
	hasher := hash.NewHashingPassword()
	tokenManager, _ := auth.NewManager("mysecret")

	userQueryClient := pb_user.NewUserQueryServiceClient(s.Conns["user"])
	userCommandClient := pb_user.NewUserCommandServiceClient(s.Conns["user"])
	roleQueryClient := pb_role.NewRoleQueryServiceClient(s.Conns["role"])
	userRoleClient := pb_user_role.NewUserRoleServiceClient(s.Conns["role"])

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, s.Log)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, s.Log)

	authRepos := auth_repo.NewRepositories(&auth_repo.Deps{
		Db:              s.ts.SQLxDB(),
		User:            userQueryClient,
		UserCommand:     userCommandClient,
		Role:            roleQueryClient,
		UserRoleCommand: userRoleClient,
		Guards: auth_repo.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	})
	authMencache := auth_cache.NewMencache(cacheStore)
	authSvc := auth_service.NewService(&auth_service.Deps{
		Repositories:  authRepos,
		Logger:        s.Log,
		Mencache:      authMencache,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: s.Obs,
	})
	authGapi := auth_handler.NewAuthHandleGrpc(authSvc, s.Log)
	server := grpc.NewServer()
	pb_auth.RegisterAuthServiceServer(server, authGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["auth"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupBannerService() {
	cacheStore := s.GetCacheStore()
	bannerMencache := banner_cache.NewMencache(cacheStore)
	bannerRepos := banner_repo.NewRepositories(s.ts.SQLxDB())
	bannerSvc := banner_service.NewService(&banner_service.Deps{
		Cache:         bannerMencache,
		Repository:    bannerRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	bannerGapi := banner_handler.NewHandler(&banner_handler.Deps{
		Service: bannerSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_banner.RegisterBannerQueryServiceServer(server, bannerGapi.BannerQuery)
	pb_banner.RegisterBannerCommandServiceServer(server, bannerGapi.BannerCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["banner"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupSliderService() {
	cacheStore := s.GetCacheStore()
	sliderMencache := slider_cache.NewMencache(cacheStore)
	sliderRepos := slider_repo.NewRepositories(s.ts.SQLxDB())
	sliderSvc := slider_service.NewService(&slider_service.Deps{
		Mencache:      sliderMencache,
		Repositories:  sliderRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	sliderGapi := slider_handler.NewHandler(&slider_handler.Deps{
		Service: sliderSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_slider.RegisterSliderQueryServiceServer(server, sliderGapi.SliderQuery)
	pb_slider.RegisterSliderCommandServiceServer(server, sliderGapi.SliderCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["slider"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCategoryService() {
	cacheStore := s.GetCacheStore()
	catMencache := category_cache.NewMencache(cacheStore)
	catRepos := category_repo.NewRepositories(s.ts.SQLxDB())
	catSvc := category_service.NewService(&category_service.Deps{
		Cache:         catMencache,
		Repositories:  catRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	catGapi := category_handler.NewHandler(&category_handler.Deps{
		Service: catSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_category.RegisterCategoryQueryServiceServer(server, catGapi.CategoryQuery)
	pb_category.RegisterCategoryCommandServiceServer(server, catGapi.CategoryCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["category"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupProductService() {
	cacheStore := s.GetCacheStore()
	prodMencache := product_cache.NewMencache(cacheStore)
	catQueryClient := pb_category.NewCategoryQueryServiceClient(s.Conns["category"])
	merchantQueryClient := pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"])
	guardCategory := resilience.NewDependencyGuard("category", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	prodRepos := product_repo.NewRepositories(s.ts.SQLxDB(), catQueryClient, merchantQueryClient,
		product_repo.GuardOptions{
			Category: []adapter.GuardOption{adapter.WithDependencyGuard(guardCategory)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		})
	prodSvc := product_service.NewService(&product_service.Deps{
		Cache:         prodMencache,
		Repository:    prodRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	prodGapi := product_handler.NewHandler(&product_handler.Deps{
		Service: prodSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_product.RegisterProductQueryServiceServer(server, prodGapi.ProductQuery)
	pb_product.RegisterProductCommandServiceServer(server, prodGapi.ProductCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["product"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCartService() {
	cacheStore := s.GetCacheStore()
	cartMencache := cart_cache.NewMencache(cacheStore)
	userQueryClient := pb_user.NewUserQueryServiceClient(s.Conns["user"])
	productQueryClient := pb_product.NewProductQueryServiceClient(s.Conns["product"])
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log)
	cartRepos := cart_repo.NewRepositories(s.ts.SQLxDB(), userQueryClient, productQueryClient,
		cart_repo.GuardOptions{
			User:    []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Product: []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
		})
	cartSvc := cart_service.NewService(&cart_service.Deps{
		Cache:         cartMencache,
		Repositories:  cartRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	cartGapi := cart_handler.NewHandler(&cart_handler.Deps{
		Service: cartSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_cart.RegisterCartQueryServiceServer(server, cartGapi.CartQuery)
	pb_cart.RegisterCartCommandServiceServer(server, cartGapi.CartCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["cart"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantService() {
	cacheStore := s.GetCacheStore()
	merchantMencache := merchant_cache.NewMencache(cacheStore)
	userQueryClient := pb_user.NewUserQueryServiceClient(s.Conns["user"])
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	merchantRepos := merchant_repo.NewRepositories(s.ts.SQLxDB(), userQueryClient,
		merchant_repo.GuardOptions{
			User: []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
		})
	merchantSvc := merchant_service.NewService(&merchant_service.Deps{
		Mencache:      merchantMencache,
		Repositories:  merchantRepos,
		Logger:        s.Log,
		Observability: s.Obs,
		Kafka:         nil,
	})
	merchantGapi := merchant_handler.NewHandler(&merchant_handler.Deps{
		Service: merchantSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_merchant.RegisterMerchantQueryServiceServer(server, merchantGapi.MerchantQuery)
	pb_merchant.RegisterMerchantCommandServiceServer(server, merchantGapi.MerchantCommandHandler)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderService() {
	cacheStore := s.GetCacheStore()
	orderMencache := order_cache.NewMencache(cacheStore)
	orderRepos := order_repo.NewRepositories(&order_repo.Deps{
		DB:                       s.ts.SQLxDB(),
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
	orderSvc := order_service.NewService(&order_service.Deps{
		Cache:         orderMencache,
		Repositories:  orderRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	orderGapi := order_handler.NewHandler(&order_handler.Deps{
		Service: orderSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_order.RegisterOrderQueryServiceServer(server, orderGapi.OrderQuery)
	pb_order.RegisterOrderCommandServiceServer(server, orderGapi.OrderCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantAwardService() {
	cacheStore := s.GetCacheStore()
	awardMencache := merchant_award_cache.NewMencache(cacheStore)
	awardRepos := merchant_award_repo.NewRepositories(s.ts.SQLxDB(),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		merchant_award_repo.GuardOptions{
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
		})
	awardSvc := merchant_award_service.NewService(&merchant_award_service.Deps{
		Cache:         awardMencache,
		Repository:    awardRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	awardGapi := merchant_award_handler.NewHandler(&merchant_award_handler.Deps{
		Service: awardSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_merchant_award.RegisterMerchantAwardQueryServiceServer(server, awardGapi.MerchantAwardQuery)
	pb_merchant_award.RegisterMerchantAwardCommandServiceServer(server, awardGapi.MerchantAwardCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant_award"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantBusinessService() {
	cacheStore := s.GetCacheStore()
	businessMencache := merchant_business_cache.NewMencache(cacheStore)
	businessRepos := merchant_business_repo.NewRepositories(s.ts.SQLxDB(),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		merchant_business_repo.GuardOptions{
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
		})
	businessSvc := merchant_business_service.NewService(&merchant_business_service.Deps{
		Cache:         businessMencache,
		Repository:    businessRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	businessGapi := merchant_business_handler.NewHandler(&merchant_business_handler.Deps{
		Service: businessSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_merchant_business.RegisterMerchantBusinessQueryServiceServer(server, businessGapi.MerchantBusinessQuery)
	pb_merchant_business.RegisterMerchantBusinessCommandServiceServer(server, businessGapi.MerchantBusinessCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant_business"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupTransactionService() {
	cacheStore := s.GetCacheStore()
	transactionMencache := transaction_cache.NewMencache(cacheStore)
	transactionRepos := transaction_repo.NewRepositories(&transaction_repo.Deps{
		DB:                   s.ts.SQLxDB(),
		UserQueryClient:      pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		MerchantQueryClient:  pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		OrderQueryClient:     pb_order.NewOrderQueryServiceClient(s.Conns["order"]),
		OrderItemQueryClient: pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		ShippingQueryClient:  pb_shipping_address.NewShippingQueryServiceClient(s.Conns["shipping-address"]),
		Guards: transaction_repo.GuardOptions{
			User:      []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			Order:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})
	transactionSvc := transaction_service.NewService(&transaction_service.Deps{
		Cache:         transactionMencache,
		Repositories:  transactionRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	transactionGapi := transaction_handler.NewHandler(&transaction_handler.Deps{
		Service: transactionSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_transaction.RegisterTransactionQueryServiceServer(server, transactionGapi.TransactionQuery)
	pb_transaction.RegisterTransactionCommandServiceServer(server, transactionGapi.TransactionCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["transaction"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantDetailService() {
	cacheStore := s.GetCacheStore()
	detailMencache := merchant_detail_cache.NewMencache(cacheStore)
	detailRepos := merchant_detail_repo.NewRepositories(s.ts.SQLxDB(),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		merchant_detail_repo.GuardOptions{
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
		})
	detailSvc := merchant_detail_service.NewService(&merchant_detail_service.Deps{
		Cache:         detailMencache,
		Repository:    detailRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	detailGapi := merchant_detail_handler.NewHandler(&merchant_detail_handler.Deps{
		Service: detailSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_merchant_detail.RegisterMerchantDetailQueryServiceServer(server, detailGapi.MerchantDetailQuery)
	pb_merchant_detail.RegisterMerchantDetailCommandServiceServer(server, detailGapi.MerchantDetailCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant_detail"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantPolicyService() {
	cacheStore := s.GetCacheStore()
	policyMencache := merchant_policy_cache.NewMencache(cacheStore)
	policyRepos := merchant_policy_repo.NewRepositories(s.ts.SQLxDB(),
		pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		merchant_policy_repo.GuardOptions{
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
		})
	policySvc := merchant_policy_service.NewService(&merchant_policy_service.Deps{
		Cache:         policyMencache,
		Repository:    policyRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	policyGapi := merchant_policy_handler.NewHandler(&merchant_policy_handler.Deps{
		Service: policySvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_merchant_policy.RegisterMerchantPolicyQueryServiceServer(server, policyGapi.MerchantPolicyQuery)
	pb_merchant_policy.RegisterMerchantPolicyCommandServiceServer(server, policyGapi.MerchantPolicyCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant_policy"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupShippingAddressService() {
	cacheStore := s.GetCacheStore()
	addrMencache := shipping_address_cache.NewMencache(cacheStore)
	addrRepos := shipping_address_repo.NewRepositories(s.ts.SQLxDB())
	addrSvc := shipping_address_service.NewService(&shipping_address_service.Deps{
		Mencache:      addrMencache,
		Repositories:  addrRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	addrGapi := shipping_address_handler.NewHandler(&shipping_address_handler.Deps{
		Service: addrSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_shipping_address.RegisterShippingQueryServiceServer(server, addrGapi.ShippingQuery)
	pb_shipping_address.RegisterShippingCommandServiceServer(server, addrGapi.ShippingCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["shipping-address"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderItemService() {
	cacheStore := s.GetCacheStore()
	itemMencache := order_item_cache.NewMencache(cacheStore)
	itemRepos := order_item_repo.NewRepositories(s.ts.SQLxDB())
	itemSvc := order_item_service.NewService(&order_item_service.Deps{
		Cache:         itemMencache,
		Repository:    itemRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	itemGapi := order_item_handler.NewHandler(&order_item_handler.Deps{
		Service: itemSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_order_item.RegisterOrderItemQueryServiceServer(server, itemGapi.OrderItemQuery)
	pb_order_item.RegisterOrderItemCommandServiceServer(server, itemGapi.OrderItemCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order-item"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupReviewService() {
	cacheStore := s.GetCacheStore()
	reviewMencache := review_cache.NewMencache(cacheStore)
	userQueryClient := pb_user.NewUserQueryServiceClient(s.Conns["user"])
	productQueryClient := pb_product.NewProductQueryServiceClient(s.Conns["product"])
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log)
	reviewRepos := review_repo.NewRepositories(s.ts.SQLxDB(), userQueryClient, productQueryClient,
		review_repo.GuardOptions{
			User:    []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Product: []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
		})
	reviewSvc := review_service.NewService(&review_service.Deps{
		Cache:         reviewMencache,
		Repositories:  reviewRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	reviewGapi := review_handler.NewHandler(&review_handler.Deps{
		Service: reviewSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_review.RegisterReviewQueryServiceServer(server, reviewGapi.ReviewQuery)
	pb_review.RegisterReviewCommandServiceServer(server, reviewGapi.ReviewCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["review"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupReviewDetailService() {
	cacheStore := s.GetCacheStore()
	detailMencache := review_detail_cache.NewMencache(cacheStore)
	detailRepos := review_detail_repo.NewRepositories(s.ts.SQLxDB())
	detailSvc := review_detail_service.NewService(&review_detail_service.Deps{
		Cache:         detailMencache,
		Repositories:  detailRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	detailGapi := review_detail_handler.NewHandler(&review_detail_handler.Deps{
		Service: detailSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pb_review_detail.RegisterReviewDetailQueryServiceServer(server, detailGapi.ReviewDetailQuery)
	pb_review_detail.RegisterReviewDetailCommandServiceServer(server, detailGapi.ReviewDetailCommand)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["review-detail"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) dial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	return conn
}

func (s *BaseTestSuite) GetCacheStore() *cache.CacheStore {
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	return cache.NewCacheStore(s.ts.RedisClient(), s.Log, cacheMetrics)
}

func (s *BaseTestSuite) BuildMultipartRequestBody(fields map[string]string, fieldName, fileName string) ([]byte, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range fields {
		fw, _ := w.CreateFormField(key)
		fw.Write([]byte(r))
	}
	fw, _ := w.CreateFormFile(fieldName, fileName)
	fw.Write([]byte("dummy image content"))
	w.Close()
	return b.Bytes(), w.FormDataContentType()
}
