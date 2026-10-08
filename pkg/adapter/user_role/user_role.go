// Package user_role adapts the UserRole service gRPC API into the shared domain
// model. It owns the only place that talks to pb/user_role's UserRoleService;
// role assignment no longer lives on the role service.
package user_role

import (
	"context"
	"time"

	pbrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pbuserrole "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	userrole_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/user_role_errors"
)

// QueryRepository is the read path consumers use to resolve a user's roles.
type QueryRepository interface {
	FindByUserId(ctx context.Context, userID int) ([]*models.Role, error)
}

// CommandRepository is the write path consumers use to (un)assign roles.
type CommandRepository interface {
	AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error)
	RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error
}

// Repository implements QueryRepository and CommandRepository on top of the
// generated user-role service client.
type Repository struct {
	client pbuserrole.UserRoleServiceClient
	guard  *resilience.DependencyGuard
}

// New wraps a generated user-role service client into a Repository. Zero or more
// guard options attach a per-call timeout, bulkhead and circuit breaker around
// every outbound gRPC call; with no options the guard stays nil and calls pass
// straight through.
func New(client pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewQueryAdapter returns the adapter restricted to the QueryRepository surface
// for consumers that only read a user's roles.
func NewQueryAdapter(client pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) QueryRepository {
	return New(client, opts...)
}

// NewCommandAdapter returns the adapter restricted to the CommandRepository
// surface for consumers (auth, user) that only (un)assign roles.
func NewCommandAdapter(client pbuserrole.UserRoleServiceClient, opts ...adapter.GuardOption) CommandRepository {
	return New(client, opts...)
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

// FindByUserId implements QueryRepository. The user-role service returns the
// resolved roles, so we map the role responses straight into the domain model.
func (r *Repository) FindByUserId(ctx context.Context, userID int) ([]*models.Role, error) {
	var res *pbrole.ApiResponsesRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.client.FindByUserId(ctx, &pbuserrole.FindByIdUserRoleRequest{UserId: int32(userID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	roles := make([]*models.Role, 0, len(res.Data))
	for _, item := range res.Data {
		if item == nil {
			continue
		}
		roles = append(roles, roleToModel(item))
	}
	return roles, nil
}

// AssignRoleToUser implements CommandRepository.
func (r *Repository) AssignRoleToUser(ctx context.Context, req *requests.CreateUserRoleRequest) (*models.UserRole, error) {
	var res *pbuserrole.ApiResponseUserRole
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = r.client.AssignRoleToUser(ctx, &pbuserrole.AssignRoleToUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil || res == nil || res.Data == nil {
		return nil, userrole_errors.ErrAssignRoleToUser.WithInternal(err)
	}
	return userRoleToModel(res.Data), nil
}

// RemoveRoleFromUser implements CommandRepository.
func (r *Repository) RemoveRoleFromUser(ctx context.Context, req *requests.RemoveUserRoleRequest) error {
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := r.client.RemoveRoleFromUser(ctx, &pbuserrole.RemoveRoleFromUserRequest{
			UserId: int32(req.UserId),
			RoleId: int32(req.RoleId),
		})
		return callErr
	})
	if err != nil {
		return userrole_errors.ErrRemoveRole.WithInternal(err)
	}
	return nil
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

func userRoleToModel(ur *pbuserrole.UserRoleResponse) *models.UserRole {
	if ur == nil {
		return nil
	}
	return &models.UserRole{
		UserRoleID: ur.UserRoleId,
		UserID:     ur.UserId,
		RoleID:     ur.RoleId,
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
