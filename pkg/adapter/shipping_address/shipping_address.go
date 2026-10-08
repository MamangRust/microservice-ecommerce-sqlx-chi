package shipping_address

import (
	"context"

	pbshipping "github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shippingaddress_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/shipping_address_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

// QueryRepository is the contract consumers depend on for shipping-address reads.
type QueryRepository interface {
	FindByID(ctx context.Context, shippingID int) (*models.ShippingAddress, error)
	FindByOrder(ctx context.Context, orderID int) (*models.ShippingAddress, error)
}

// CommandRepository is the contract consumers depend on for shipping-address writes.
type CommandRepository interface {
	Create(ctx context.Context, req *requests.CreateShippingAddressRequest) (*models.ShippingAddress, error)
	Update(ctx context.Context, req *requests.UpdateShippingAddressRequest) (*models.ShippingAddress, error)
	DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}

// Repository implements both QueryRepository and CommandRepository on top of the
// generated shipping query/command clients.
type Repository struct {
	query   pbshipping.ShippingQueryServiceClient
	command pbshipping.ShippingCommandServiceClient
	guard   *resilience.DependencyGuard
}

// NewAdapter wraps the generated shipping clients. Zero or more guard options
// attach a per-call timeout, bulkhead and circuit breaker around every outbound
// gRPC call; with no options the guard stays nil and calls pass straight
// through.
func NewAdapter(query pbshipping.ShippingQueryServiceClient, command pbshipping.ShippingCommandServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{query: query, command: command}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewQueryAdapter returns the adapter restricted to the QueryRepository surface
// for consumers that only read.
func NewQueryAdapter(query pbshipping.ShippingQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	r := &Repository{query: query}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewCommandAdapter returns the adapter restricted to the CommandRepository
// surface for consumers that only write.
func NewCommandAdapter(command pbshipping.ShippingCommandServiceClient, opts ...adapter.GuardOption) CommandRepository {
	r := &Repository{command: command}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (a *Repository) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

// Reads propagate the dependency's gRPC status unchanged (NotFound -> 404, and
// so on). Command failures keep the domain-error wrapping the order service
// already relied on; AppError.Unwrap still exposes the underlying gRPC status.
func (a *Repository) FindByID(ctx context.Context, shippingID int) (*models.ShippingAddress, error) {
	var res *pbshipping.ApiResponseShipping
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.query.FindById(ctx, &pbshipping.FindByIdShippingRequest{Id: int32(shippingID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return toModel(res.Data), nil
}

func (a *Repository) FindByOrder(ctx context.Context, orderID int) (*models.ShippingAddress, error) {
	var res *pbshipping.ApiResponseShipping
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.query.FindByOrder(ctx, &pbshipping.FindByIdShippingRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return toModel(res.Data), nil
}

func (a *Repository) Create(ctx context.Context, req *requests.CreateShippingAddressRequest) (*models.ShippingAddress, error) {
	var orderID int32
	if req.OrderID != nil {
		orderID = int32(*req.OrderID)
	}

	var res *pbshipping.ApiResponseShipping
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.CreateShipping(ctx, &pbshipping.CreateShippingAddressRequest{
			OrderId:        orderID,
			Alamat:         req.Alamat,
			Provinsi:       req.Provinsi,
			Kota:           req.Kota,
			Negara:         req.Negara,
			Courier:        req.Courier,
			ShippingMethod: req.ShippingMethod,
			ShippingCost:   int32(req.ShippingCost),
		})
		return callErr
	})
	if err != nil {
		return nil, shippingaddress_errors.ErrCreateShippingAddress.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) Update(ctx context.Context, req *requests.UpdateShippingAddressRequest) (*models.ShippingAddress, error) {
	var shippingID int32
	if req.ShippingID != nil {
		shippingID = int32(*req.ShippingID)
	}

	var res *pbshipping.ApiResponseShipping
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.UpdateShipping(ctx, &pbshipping.UpdateShippingAddressRequest{
			ShippingId:     shippingID,
			Alamat:         req.Alamat,
			Provinsi:       req.Provinsi,
			Kota:           req.Kota,
			Negara:         req.Negara,
			Courier:        req.Courier,
			ShippingMethod: req.ShippingMethod,
			ShippingCost:   int32(req.ShippingCost),
		})
		return callErr
	})
	if err != nil {
		return nil, shippingaddress_errors.ErrUpdateShippingAddress.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error) {
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := a.command.DeleteShippingByOrderPermanent(ctx, &pbshipping.FindByIdShippingRequest{Id: int32(orderID)})
		return callErr
	})
	if err != nil {
		return false, shippingaddress_errors.ErrDeleteShippingAddressPermanent.WithInternal(err)
	}
	return true, nil
}

func (a *Repository) DeleteAll(ctx context.Context) (bool, error) {
	var res *pbshipping.ApiResponseShippingAll
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.DeleteAllShippingPermanent(ctx, &emptypb.Empty{})
		return callErr
	})
	if err != nil {
		return false, shippingaddress_errors.ErrDeleteAllPermanentShippingAddress.WithInternal(err)
	}
	return res.Status == "success", nil
}

func toModel(data *pbshipping.ShippingResponse) *models.ShippingAddress {
	if data == nil {
		return nil
	}
	return &models.ShippingAddress{
		ShippingAddressID: data.Id,
		OrderID:           data.OrderId,
		Alamat:            data.Alamat,
		Provinsi:          data.Provinsi,
		Negara:            data.Negara,
		Kota:              data.Kota,
		Courier:           data.Courier,
		ShippingMethod:    data.ShippingMethod,
		ShippingCost:      float64(data.ShippingCost),
	}
}
