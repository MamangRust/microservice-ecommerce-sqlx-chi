package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	RoleCommand RoleCommandRepository
	RoleQuery   RoleQueryRepository
	UserRole    UserRoleRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		RoleCommand: NewRoleCommandRepository(DB),
		RoleQuery:   NewRoleQueryRepository(DB),
		UserRole:    NewUserRoleRepository(DB),
	}
}
