package repository

import (
	pb_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	pb_user_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/role"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	userroleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user_role"
	"github.com/jmoiron/sqlx"
)

// Repositories assembles auth's own stores (refresh/reset tokens) with the
// shared gRPC adapters for the identity tables auth no longer owns (users,
// roles, user_roles). auth therefore never issues SQL against those tables, and
// the adapters are consumed directly rather than wrapped by a repository struct.
type Repositories struct {
	User         UserRepository
	UserCommand  useradapter.CommandRepository
	Role         roleadapter.QueryRepository
	UserRole     userroleadapter.CommandRepository
	RefreshToken RefreshTokenRepository
	ResetToken   ResetTokenRepository
}

// GuardOptions carries the resilience guard options for each outbound
// dependency. Callers build them with adapter.WithDependencyGuard.
type GuardOptions struct {
	User     []adapter.GuardOption
	UserRole []adapter.GuardOption
	Role     []adapter.GuardOption
}

type Deps struct {
	Db              *sqlx.DB
	User            pb_user.UserQueryServiceClient
	UserCommand     pb_user.UserCommandServiceClient
	Role            pb_role.RoleQueryServiceClient
	UserRoleCommand pb_user_role.UserRoleServiceClient
	Guards          GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	userAdapter := useradapter.NewAdapter(deps.User, deps.UserCommand, deps.Guards.User...)

	return &Repositories{
		User:         userAdapter,
		UserCommand:  userAdapter,
		Role:         roleadapter.NewQueryAdapter(deps.Role, deps.Guards.Role...),
		UserRole:     userroleadapter.NewCommandAdapter(deps.UserRoleCommand, deps.Guards.UserRole...),
		RefreshToken: NewRefreshTokenRepository(deps.Db),
		ResetToken:   NewResetTokenRepository(deps.Db),
	}
}
