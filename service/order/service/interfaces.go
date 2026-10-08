package service

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

// F5: legacy OLTP order stats services were removed; stats are served by
// service/stats_reader from ClickHouse.

type OrderQueryService interface {
	FindAll(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersRow, *int, error)

	FindActive(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersActiveRow, *int, error)

	FindTrashed(
		ctx context.Context,
		req *requests.FindAllOrder,
	) ([]*db.GetOrdersTrashedRow, *int, error)

	FindByID(
		ctx context.Context,
		order_id int,
	) (*db.GetOrderByIDRow, error)
}

type OrderCommandService interface {
	Create(
		ctx context.Context,
		request *requests.CreateOrderRequest,
	) (*db.CreateOrderRow, error)

	Update(
		ctx context.Context,
		request *requests.UpdateOrderRequest,
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
	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)

	// ReconcileStockReservations repairs drift between reservation status and order
	// lifecycle caused by failed compensation (e.g. released reservation on an active
	// order, or reserved reservation on a trashed order). It returns the number of
	// reservations repaired.
	ReconcileStockReservations(ctx context.Context) (*ReconcileResult, error)

	// CleanupIdempotencyRecords applies the retention policy to the idempotency ledger
	// and to released reservations of trashed orders older than the retention window.
	CleanupIdempotencyRecords(ctx context.Context, retentionDays int) (*CleanupResult, error)
}
