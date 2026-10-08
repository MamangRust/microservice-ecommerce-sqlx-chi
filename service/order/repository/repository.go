package repository

import (
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	pb_order_item "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	pb_product "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	pb_shipping_address "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	pb_transaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	shippingadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/shipping_address"
	transactionadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/transaction"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	User        []adapter.GuardOption
	Product     []adapter.GuardOption
	Merchant    []adapter.GuardOption
	OrderItem   []adapter.GuardOption
	Shipping    []adapter.GuardOption
	Transaction []adapter.GuardOption
}

type Repositories struct {
	OrderQuery         OrderQueryRepository
	OrderCommand       OrderCommandRepository
	UserQuery          useradapter.QueryRepository
	ProductQuery       productadapter.QueryRepository
	ProductCommand     productadapter.CommandRepository
	OrderItemQuery     orderitemadapter.QueryRepository
	OrderItemCommand   orderitemadapter.CommandRepository
	MerchantQuery      merchantadapter.QueryRepository
	ShippingAddress    shippingadapter.CommandRepository
	TransactionCommand transactionadapter.CommandRepository
	ShippingQuery      shippingadapter.QueryRepository
	StockReservation   StockReservationRepository
	Outbox             OutboxRepository
}

type Deps struct {
	DB *sqlx.DB

	UserQueryClient          pb_user.UserQueryServiceClient
	ProductQueryClient       pb_product.ProductQueryServiceClient
	ProductCommandClient     pb_product.ProductCommandServiceClient
	MerchantQueryClient      pb_merchant.MerchantQueryServiceClient
	OrderItemQueryClient     pb_order_item.OrderItemQueryServiceClient
	OrderItemCommandClient   pb_order_item.OrderItemCommandServiceClient
	ShippingCommandClient    pb_shipping_address.ShippingCommandServiceClient
	ShippingQueryClient      pb_shipping_address.ShippingQueryServiceClient
	TransactionCommandClient pb_transaction.TransactionCommandServiceClient

	Guards GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	g := deps.Guards

	productAdapter := productadapter.NewAdapter(deps.ProductQueryClient, deps.ProductCommandClient, g.Product...)
	orderItemAdapter := orderitemadapter.NewAdapter(deps.OrderItemQueryClient, deps.OrderItemCommandClient, g.OrderItem...)
	shippingAdapter := shippingadapter.NewAdapter(deps.ShippingQueryClient, deps.ShippingCommandClient, g.Shipping...)

	return &Repositories{
		OrderQuery:         NewOrderQueryRepository(deps.DB),
		OrderCommand:       NewOrderCommandRepository(deps.DB),
		UserQuery:          useradapter.NewQueryAdapter(deps.UserQueryClient, g.User...),
		ProductQuery:       productAdapter,
		ProductCommand:     productAdapter,
		OrderItemQuery:     orderItemAdapter,
		OrderItemCommand:   orderItemAdapter,
		MerchantQuery:      merchantadapter.NewQueryAdapter(deps.MerchantQueryClient, g.Merchant...),
		ShippingAddress:    shippingAdapter,
		TransactionCommand: transactionadapter.NewCommandAdapter(deps.TransactionCommandClient, g.Transaction...),
		ShippingQuery:      shippingAdapter,
		StockReservation:   NewStockReservationRepository(deps.DB),
		Outbox:             NewOutboxRepository(deps.DB),
	}
}
