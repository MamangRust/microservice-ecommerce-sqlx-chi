package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-auth/database/schema"
	roleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/role"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	userroleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user_role"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/jmoiron/sqlx"
)

// UserRepository is the auth-facing contract for user reads/writes, satisfied
// by the shared user adapter. Auth needs both surfaces, so it composes them.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type UserRepository interface {
	useradapter.QueryRepository
	useradapter.CommandRepository
}

type ResetTokenRepository interface {
	FindByToken(ctx context.Context, code string) (*db.ResetToken, error)

	CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.ResetToken, error)

	CreateResetTokenInTx(ctx context.Context, tx *sqlx.Tx, req *requests.CreateResetTokenRequest) (*db.ResetToken, error)

	DeleteResetToken(ctx context.Context, user_id int) error
}

type RefreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*db.RefreshToken, error)

	FindByUserId(ctx context.Context, user_id int) (*db.RefreshToken, error)

	CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*db.RefreshToken, error)

	UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*db.RefreshToken, error)

	DeleteRefreshToken(ctx context.Context, token string) error

	DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error
}

// UserRoleRepository is satisfied by the user_role adapter; role assignment now
// lives in the user_role service, not the role service.
type UserRoleRepository = userroleadapter.CommandRepository

type RoleRepository = roleadapter.QueryRepository
