package repository

import (
	"context"

	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	shippingadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/shipping_address"
	transactionadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/transaction"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

type UserQueryRepository = useradapter.QueryRepository

type ProductQueryRepository = productadapter.QueryRepository

type MerchantQueryRepository = merchantadapter.QueryRepository

type ProductCommandRepository = productadapter.CommandRepository

type ShippingAddressCommandRepository = shippingadapter.CommandRepository

type TransactionCommandRepository = transactionadapter.CommandRepository

type OrderItemQueryRepository = orderitemadapter.QueryRepository

type OrderItemCommandRepository = orderitemadapter.CommandRepository

// F5: legacy OLTP order stats repositories were removed; stats are served by
// service/stats_reader from ClickHouse.

type OrderCommandRepository interface {
	Create(
		ctx context.Context,
		request *requests.CreateOrderRecordRequest,
	) (*db.CreateOrderRow, error)

	Update(
		ctx context.Context,
		request *requests.UpdateOrderRecordRequest,
	) (*db.UpdateOrderRow, error)

	Trash(
		ctx context.Context,
		order_id int,
	) (*db.Order, error)

	Restore(
		ctx context.Context,
		order_id int,
	) (*db.Order, error)

	DeletePermanent(
		ctx context.Context,
		order_id int,
	) (bool, error)

	// DeletePermanentWithChildren atomically removes a trashed order and all of
	// its child rows (stock reservations, order items, transactions, shipping
	// addresses) in a single SQL statement. It returns ErrOrderNotFound when the
	// order is not trashed.
	DeletePermanentWithChildren(
		ctx context.Context,
		order_id int,
	) (bool, error)
	FindTrashedByID(ctx context.Context, order_id int) (*db.Order, error)
	FindTrashed(ctx context.Context) ([]*db.Order, error)

	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}

type OrderQueryRepository interface {
	FindAll(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersRow, error)

	FindActive(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersActiveRow, error)

	FindTrashed(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersTrashedRow, error)

	FindByMerchant(
		ctx context.Context,
		req *requests.FindAllOrderByMerchant,
	) ([]*db.GetOrdersByMerchantRow, error)

	FindByID(
		ctx context.Context,
		order_id int,
	) (*db.GetOrderByIDRow, error)
}
