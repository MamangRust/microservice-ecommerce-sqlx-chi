// Package role adapts the Role query service gRPC API into the shared domain
// model. Role assignment lives in the user_role service, so this adapter only
// exposes the read path.
package role

import (
	"context"
	"time"

	pbrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/role_errors"
)

// QueryRepository is the contract consumers depend on for role reads.
type QueryRepository interface {
	FindById(ctx context.Context, roleID int) (*models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

type grpcQueryAdapter struct {
	client pbrole.RoleQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewQueryAdapter wraps a generated role query client into a QueryRepository.
// Zero or more guard options attach a per-call timeout, bulkhead and circuit
// breaker around every outbound gRPC call; with no options the guard stays nil
// and calls pass straight through.
func NewQueryAdapter(client pbrole.RoleQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &grpcQueryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *grpcQueryAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *grpcQueryAdapter) FindById(ctx context.Context, roleID int) (*models.Role, error) {
	var res *pbrole.ApiResponseRole
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindByIdRole(ctx, &pbrole.FindByIdRoleRequest{RoleId: int32(roleID)})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return roleToModel(res.Data), nil
}

func (a *grpcQueryAdapter) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var res *pbrole.ApiResponseRole
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindByNameRole(ctx, &pbrole.FindByNameRoleRequest{Name: name})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, role_errors.ErrRoleNotFound.WithInternal(err)
	}
	return roleToModel(res.Data), nil
}

func roleToModel(role *pbrole.RoleResponse) *models.Role {
	if role == nil {
		return nil
	}
	return &models.Role{
		RoleID:    role.Id,
		RoleName:  role.Name,
		CreatedAt: parseTimeString(role.CreatedAt),
		UpdatedAt: parseTimeString(role.UpdatedAt),
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
