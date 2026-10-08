package order_item

import (
	"context"
	"time"

	pborder_item "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	order_item_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/order_item_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

// QueryRepository is the contract consumers depend on for order-item reads.
type QueryRepository interface {
	FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error)
	// CalculateTotalPrice is served by the command service (it is the RPC the
	// owning service exposes), but callers treat it as a read.
	CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error)
}

// BulkRepository enumerates order items page by page for the stats backfill.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.OrderItem, int, error)
}

// CommandRepository is the contract consumers depend on for order-item writes.
type CommandRepository interface {
	Create(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error)
	Update(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error)
	Trash(ctx context.Context, orderID int) (*models.OrderItem, error)
	Restore(ctx context.Context, orderID int) (*models.OrderItem, error)
	DeletePermanent(ctx context.Context, orderID int) (bool, error)
	DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error)
	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}

// Repository implements both QueryRepository and CommandRepository on top of the
// generated order-item query/command clients.
type Repository struct {
	query   pborder_item.OrderItemQueryServiceClient
	command pborder_item.OrderItemCommandServiceClient
	guard   *resilience.DependencyGuard
}

// NewAdapter wraps the generated order-item clients. Zero or more guard options
// attach a per-call timeout, bulkhead and circuit breaker around every outbound
// gRPC call; with no options the guard stays nil and calls pass straight
// through.
func NewAdapter(query pborder_item.OrderItemQueryServiceClient, command pborder_item.OrderItemCommandServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{query: query, command: command}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewQueryAdapter returns the adapter restricted to the QueryRepository surface
// for consumers that only read.
func NewQueryAdapter(query pborder_item.OrderItemQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	r := &Repository{query: query}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewBulkAdapter returns the adapter restricted to the BulkRepository surface for
// the stats backfill.
func NewBulkAdapter(query pborder_item.OrderItemQueryServiceClient, opts ...adapter.GuardOption) BulkRepository {
	r := &Repository{query: query}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (a *Repository) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *Repository) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	var res *pborder_item.ApiResponsesOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.query.FindOrderItemByOrder(ctx, &pborder_item.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrFindOrderItemByOrder.WithInternal(err)
	}

	items := make([]*models.OrderItem, 0, len(res.Data))
	for _, item := range res.Data {
		items = append(items, toModel(item))
	}

	return items, nil
}

func (a *Repository) CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error) {
	var res *pborder_item.CalculateTotalPriceResponse
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.CalculateTotalPrice(ctx, &pborder_item.CalculateTotalPriceRequest{OrderId: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrCalculateTotalPrice.WithInternal(err)
	}

	total := int32(res.TotalPrice)
	return &total, nil
}

func (a *Repository) Create(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	var res *pborder_item.ApiResponseOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.CreateOrderItem(ctx, &pborder_item.CreateOrderItemRecordRequest{
			OrderId:   int32(req.OrderID),
			ProductId: int32(req.ProductID),
			Quantity:  int32(req.Quantity),
			Price:     int32(req.Price),
		})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrCreateOrderItem.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) Update(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	var res *pborder_item.ApiResponseOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.UpdateOrderItem(ctx, &pborder_item.UpdateOrderItemRecordRequest{
			OrderItemId: int32(req.OrderItemID),
			Quantity:    int32(req.Quantity),
			Price:       int32(req.Price),
		})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrUpdateOrderItem.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) Trash(ctx context.Context, orderID int) (*models.OrderItem, error) {
	var res *pborder_item.ApiResponseOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.TrashOrderItem(ctx, &pborder_item.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrTrashedOrderItem.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) Restore(ctx context.Context, orderID int) (*models.OrderItem, error) {
	var res *pborder_item.ApiResponseOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.RestoreOrderItem(ctx, &pborder_item.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, order_item_errors.ErrRestoreOrderItem.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) DeletePermanent(ctx context.Context, orderID int) (bool, error) {
	var res *pborder_item.ApiResponseOrderItemDelete
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.DeleteOrderItemPermanent(ctx, &pborder_item.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return false, order_item_errors.ErrDeleteOrderItemPermanent.WithInternal(err)
	}
	return res.Status == "success", nil
}

func (a *Repository) DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error) {
	var res *pborder_item.ApiResponseOrderItemDelete
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.DeleteOrderItemByOrderPermanent(ctx, &pborder_item.FindByIdOrderItemRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return false, order_item_errors.ErrDeleteOrderItemPermanent.WithInternal(err)
	}
	return res.Status == "success", nil
}

func (a *Repository) RestoreAll(ctx context.Context) (bool, error) {
	var res *pborder_item.ApiResponseOrderItemAll
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.RestoreAllOrdersItem(ctx, &emptypb.Empty{})
		return callErr
	})
	if err != nil {
		return false, order_item_errors.ErrRestoreAllOrderItem.WithInternal(err)
	}
	return res.Status == "success", nil
}

func (a *Repository) DeleteAll(ctx context.Context) (bool, error) {
	var res *pborder_item.ApiResponseOrderItemAll
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.DeleteAllPermanentOrdersItem(ctx, &emptypb.Empty{})
		return callErr
	})
	if err != nil {
		return false, order_item_errors.ErrDeleteAllOrderPermanent.WithInternal(err)
	}
	return res.Status == "success", nil
}

func (a *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.OrderItem, int, error) {
	var resp *pborder_item.ApiResponsePaginationOrderItem
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = a.query.FindAll(ctx, &pborder_item.FindAllOrderItemRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil {
		return nil, 0, order_item_errors.ErrFindAllOrderItems.WithInternal(err)
	}
	items := make([]*models.OrderItem, 0, len(resp.Data))
	for _, it := range resp.Data {
		items = append(items, toModel(it))
	}
	total := len(items)
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return items, total, nil
}

func toModel(item *pborder_item.OrderItemResponse) *models.OrderItem {
	return &models.OrderItem{
		OrderItemID: item.Id,
		OrderID:     item.OrderId,
		ProductID:   item.ProductId,
		Quantity:    item.Quantity,
		Price:       item.Price,
		CreatedAt:   parseTime(item.CreatedAt),
	}
}

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
