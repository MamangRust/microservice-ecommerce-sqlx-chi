package repository

import (
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	pb_order "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	pb_order_item "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	pb_shipping_address "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	orderadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	shippingadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/shipping_address"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	TransactionCommand TransactionCommandRepository
	TransactionQuery   TransactionQueryRepository
	OrderItem          orderitemadapter.QueryRepository
	OrderQuery         orderadapter.QueryRepository
	MerchantQuery      merchantadapter.QueryRepository
	ShippingAddress    shippingadapter.QueryRepository
	UserQuery          useradapter.QueryRepository
	Outbox             OutboxRepository
}

// GuardOptions carries the resilience guard options for each outbound
// dependency.
type GuardOptions struct {
	User      []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Order     []adapter.GuardOption
	OrderItem []adapter.GuardOption
	Shipping  []adapter.GuardOption
}

type Deps struct {
	DB *sqlx.DB

	UserQueryClient      pb_user.UserQueryServiceClient
	MerchantQueryClient  pb_merchant.MerchantQueryServiceClient
	OrderQueryClient     pb_order.OrderQueryServiceClient
	OrderItemQueryClient pb_order_item.OrderItemQueryServiceClient
	ShippingQueryClient  pb_shipping_address.ShippingQueryServiceClient

	Guards GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	g := deps.Guards

	return &Repositories{
		TransactionCommand: NewTransactionCommandRepository(deps.DB),
		TransactionQuery:   NewTransactionQueryRepository(deps.DB),
		OrderItem:          orderitemadapter.NewQueryAdapter(deps.OrderItemQueryClient, g.OrderItem...),
		OrderQuery:         orderadapter.NewQueryAdapter(deps.OrderQueryClient, g.Order...),
		MerchantQuery:      merchantadapter.NewQueryAdapter(deps.MerchantQueryClient, g.Merchant...),
		ShippingAddress:    shippingadapter.NewQueryAdapter(deps.ShippingQueryClient, g.Shipping...),
		UserQuery:          useradapter.NewQueryAdapter(deps.UserQueryClient, g.User...),
		Outbox:             NewOutboxRepository(deps.DB),
	}
}
