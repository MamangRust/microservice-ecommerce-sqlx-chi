package repository

import (
	pb_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	pb_user_role "github.com/MamangRust/microservice-ecommerce-grpc-pb/user_role"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user_role"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

type Repositories struct {
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        roleadapter.QueryRepository
	UserRole    userroleadapter.CommandRepository
}

type Deps struct {
	Db       *sqlx.DB
	Role     pb_role.RoleQueryServiceClient
	UserRole pb_user_role.UserRoleServiceClient
	Guards   GuardOptions
}

func NewRepositories(deps *Deps) *Repositories {
	return &Repositories{
		UserCommand: NewUserCommandRepository(deps.Db),
		UserQuery:   NewUserQueryRepository(deps.Db),
		Role:        roleadapter.NewQueryAdapter(deps.Role, deps.Guards.Role...),
		UserRole:    userroleadapter.NewCommandAdapter(deps.UserRole, deps.Guards.UserRole...),
	}
}
