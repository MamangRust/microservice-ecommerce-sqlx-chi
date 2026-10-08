package order

import (
	"context"
	"time"

	pborder "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
)

// QueryRepository is the contract consumers depend on for order reads.
type QueryRepository interface {
	FindByID(ctx context.Context, orderID int) (*models.Order, error)
}

// BulkRepository enumerates orders page by page. It is used by the stats
// backfill, which has to walk every historical order to materialize ClickHouse
// aggregates. The page size is supplied by the caller (the backfill walks in
// chunks); the returned int is the total record count reported by the service.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Order, int, error)
}

type queryAdapter struct {
	client pborder.OrderQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewQueryAdapter wraps the generated order query client. Zero or more guard
// options attach a per-call timeout, bulkhead and circuit breaker around every
// outbound gRPC call; with no options the guard stays nil and calls pass
// straight through.
func NewQueryAdapter(client pborder.OrderQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &queryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// NewBulkAdapter wraps the generated order query client for bulk enumeration.
func NewBulkAdapter(client pborder.OrderQueryServiceClient, opts ...adapter.GuardOption) BulkRepository {
	a := &queryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *queryAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// FindByID propagates the dependency's gRPC status unchanged (NotFound -> 404,
// and so on) instead of wrapping it in a domain error.
func (a *queryAdapter) FindByID(ctx context.Context, orderID int) (*models.Order, error) {
	var res *pborder.ApiResponseOrder
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindById(ctx, &pborder.FindByIdOrderRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return toModel(res.Data), nil
}

func (a *queryAdapter) FindAll(ctx context.Context, page, pageSize int) ([]*models.Order, int, error) {
	var resp *pborder.ApiResponsePaginationOrder
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = a.client.FindAll(ctx, &pborder.FindAllOrderRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil {
		return nil, 0, err
	}
	orders := make([]*models.Order, 0, len(resp.Data))
	for _, o := range resp.Data {
		orders = append(orders, toModel(o))
	}
	total := len(orders)
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return orders, total, nil
}

func toModel(o *pborder.OrderResponse) *models.Order {
	if o == nil {
		return nil
	}
	return &models.Order{
		OrderID:    o.Id,
		UserID:     o.UserId,
		MerchantID: o.MerchantId,
		TotalPrice: o.TotalPrice,
		CreatedAt:  parseTime(o.CreatedAt),
	}
}

// parseTime accepts the loose RFC3339-ish timestamp strings the order service
// emits and returns nil for empty strings.
func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
