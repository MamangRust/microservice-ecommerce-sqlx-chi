package merchant

import (
	"context"
	"time"

	pbmerchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
)

// QueryRepository is the contract consumers depend on for merchant reads.
// It is implemented by grpcQueryAdapter and can be satisfied by a test double.
type QueryRepository interface {
	FindByID(ctx context.Context, merchantID int) (*models.Merchant, error)
}

type grpcQueryAdapter struct {
	client pbmerchant.MerchantQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewQueryAdapter wraps a generated merchant query client into a QueryRepository.
// Zero or more guard options attach a per-call timeout, bulkhead and circuit
// breaker around every outbound gRPC call; with no options the guard stays nil
// and calls pass straight through.
func NewQueryAdapter(client pbmerchant.MerchantQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &grpcQueryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *grpcQueryAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *grpcQueryAdapter) FindByID(ctx context.Context, merchantID int) (*models.Merchant, error) {
	var res *pbmerchant.ApiResponseMerchant
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: int32(merchantID)})
		return callErr
	})
	if err != nil {
		return nil, merchant_errors.ErrMerchantNotFound.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func toModel(m *pbmerchant.MerchantResponse) *models.Merchant {
	if m == nil {
		return nil
	}
	return &models.Merchant{
		MerchantID:   m.Id,
		UserID:       m.UserId,
		Name:         m.Name,
		Description:  &m.Description,
		Address:      &m.Address,
		ContactEmail: &m.ContactEmail,
		ContactPhone: &m.ContactPhone,
		Status:       m.Status,
		CreatedAt:    parseTimeString(m.CreatedAt),
		UpdatedAt:    parseTimeString(m.UpdatedAt),
	}
}

func parseTimeString(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05.000", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
		if err != nil {
			return nil
		}
	}
	return &t
}
