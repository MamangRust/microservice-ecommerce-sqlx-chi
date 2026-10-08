// Package db holds the database row models for the role service.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Role struct {
	RoleID    int32            `json:"role_id" db:"role_id"`
	RoleName  string           `json:"role_name" db:"role_name"`
	CreatedAt pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32            `json:"user_role_id" db:"user_role_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	RoleID     int32            `json:"role_id" db:"role_id"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
