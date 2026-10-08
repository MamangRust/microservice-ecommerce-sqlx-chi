package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	apicache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache"
	auth_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/banner"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/cart"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_award"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_business"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_document"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/merchant_social_link"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/review_detail"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/auth"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

// ServiceConnections aggregates gRPC connections to backend services.
type ServiceConnections struct {
	Auth             *grpc.ClientConn
	Role             *grpc.ClientConn
	User             *grpc.ClientConn
	Category         *grpc.ClientConn
	Merchant         *grpc.ClientConn
	OrderItem        *grpc.ClientConn
	Order            *grpc.ClientConn
	Product          *grpc.ClientConn
	Transaction      *grpc.ClientConn
	Cart             *grpc.ClientConn
	Review           *grpc.ClientConn
	Slider           *grpc.ClientConn
	Shipping         *grpc.ClientConn
	Banner           *grpc.ClientConn
	MerchantAward    *grpc.ClientConn
	MerchantBusiness *grpc.ClientConn
	MerchantDetail   *grpc.ClientConn
	MerchantDocument *grpc.ClientConn
	MerchantSocial   *grpc.ClientConn
	MerchantPolicy   *grpc.ClientConn
	ReviewDetail     *grpc.ClientConn
	Card             *grpc.ClientConn
	Saldo            *grpc.ClientConn
	Topup            *grpc.ClientConn
	Transfer         *grpc.ClientConn
	Withdraw         *grpc.ClientConn
}

type Deps struct {
	Router             chi.Router
	Logger             logger.LoggerInterface
	ServiceConnections *ServiceConnections
	Cache              *cache.CacheStore
	Image              upload_image.ImageUploads
	Kafka              *kafka.Kafka
	Token              auth.TokenManager
}

func NewHandler(deps *Deps) {
	observability, _ := observability.NewObservability("apigateway", deps.Logger)
	apiHandler := apierror.NewApiHandler(observability, deps.Logger)
	authhandler.RegisterAuthHandler(&authhandler.DepsAuth{
		Client:     deps.ServiceConnections.Auth,
		Router:     deps.Router,
		Logger:     deps.Logger,
		Cache:      auth_cache.NewMencache(deps.Cache),
		ApiHandler: apiHandler,
	})

	bannerhandler.RegisterBannerHandler(&bannerhandler.DepsBanner{
		Client:        deps.ServiceConnections.Banner,
		Router:        deps.Router,
		Logger:        deps.Logger,
		CacheStore:    deps.Cache,
		Observability: observability,
	})

	carthandler.RegisterCartHandler(&carthandler.DepsCart{
		Client:     deps.ServiceConnections.Cart,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
	})

	categoryhandler.RegisterCategoryHandler(&categoryhandler.DepsCategory{
		Client:      deps.ServiceConnections.Category,
		Router:      deps.Router,
		Logger:      deps.Logger,
		CacheStore:  deps.Cache,
		UploadImage: deps.Image,
		ApiHandler:  apiHandler,
	})

	merchanthandler.RegisterMerchantHandler(&merchanthandler.DepsMerchant{
		Client:      deps.ServiceConnections.Merchant,
		Router:      deps.Router,
		Logger:      deps.Logger,
		CacheStore:  deps.Cache,
		UploadImage: deps.Image,
		ApiHandler:  apiHandler,
	})

	merchantawardhandler.RegisterMerchantAwardHandler(&merchantawardhandler.DepsMerchantAward{
		Client:     deps.ServiceConnections.MerchantAward,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
	})

	merchantbusinesshandler.RegisterMerchantBusinessHandler(&merchantbusinesshandler.DepsMerchantBusiness{
		Client:     deps.ServiceConnections.MerchantBusiness,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
	})

	merchantdocumenthandler.RegisterMerchantDocumentHandler(&merchantdocumenthandler.DepsMerchantDocument{
		Client:      deps.ServiceConnections.MerchantDocument,
		Router:      deps.Router,
		Logger:      deps.Logger,
		UploadImage: deps.Image,
	})

	merchantpolicyhandler.RegisterMerchantPolicyHandler(&merchantpolicyhandler.DepsMerchantPolicy{
		Client:        deps.ServiceConnections.MerchantPolicy,
		Router:        deps.Router,
		Logger:        deps.Logger,
		CacheStore:    deps.Cache,
		Observability: observability,
	})

	merchantsociallinkhandler.RegisterMerchantSocialLinkHandler(&merchantsociallinkhandler.DepsMerchantSocialLink{
		Client: deps.ServiceConnections.MerchantSocial,
		Router: deps.Router,
		Logger: deps.Logger,
	})

	orderhandler.RegisterOrderHandler(&orderhandler.DepsOrder{
		Client:     deps.ServiceConnections.Order,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
	})

	orderitemhandler.RegisterOrderItemHandler(&orderitemhandler.DepsOrderItem{
		Client:     deps.ServiceConnections.OrderItem,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
	})

	producthandler.RegisterProductHandler(&producthandler.DepsProduct{
		Client:     deps.ServiceConnections.Product,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
		Upload:     deps.Image,
		ApiHandler: apiHandler,
	})

	transactionhandler.RegisterTransactionHandler(&transactionhandler.DepsTransaction{
		Client:     deps.ServiceConnections.Transaction,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
		ApiHandler: apiHandler,
	})

	merchantdetailhandler.RegisterMerchantDetailHandler(&merchantdetailhandler.DepsMerchantDetail{
		Client:      deps.ServiceConnections.MerchantDetail,
		Router:      deps.Router,
		Logger:      deps.Logger,
		CacheStore:  deps.Cache,
		UploadImage: deps.Image,
		ApiHandler:  apiHandler,
	})

	rolehandler.RegisterRoleHandler(&rolehandler.DepsRole{
		Client:     deps.ServiceConnections.Role,
		Router:     deps.Router,
		Logger:     deps.Logger,
		CacheStore: deps.Cache,
		Cache:      apicache.NewRoleCache(deps.Cache),
		ApiHandler: apiHandler,
	})

	sliderhandler.RegisterSliderHandler(&sliderhandler.DepsSlider{
		Client: deps.ServiceConnections.Slider,
		Router: deps.Router,
		Logger: deps.Logger,
		Cache:  deps.Cache,
		Upload: deps.Image,
	})

	reviewhandler.RegisterReviewHandler(&reviewhandler.DepsReview{
		Client:        deps.ServiceConnections.Review,
		Router:        deps.Router,
		Logger:        deps.Logger,
		Cache:         deps.Cache,
		Observability: observability,
	})

	reviewdetailhandler.RegisterReviewDetailHandler(&reviewdetailhandler.DepsReviewDetail{
		Client:        deps.ServiceConnections.ReviewDetail,
		Router:        deps.Router,
		Logger:        deps.Logger,
		Cache:         deps.Cache,
		Upload:        deps.Image,
		Observability: observability,
	})

	shippingaddresshandler.RegisterShippingAddressHandler(&shippingaddresshandler.DepsShippingAddress{
		Client: deps.ServiceConnections.Shipping,
		Router: deps.Router,
		Logger: deps.Logger,
		Cache:  deps.Cache,
	})

	userhandler.RegisterUserHandler(&userhandler.DepsUser{
		Client:     deps.ServiceConnections.User,
		Router:     deps.Router,
		Logger:     deps.Logger,
		Cache:      deps.Cache,
		ApiHandler: apiHandler,
	})
}
