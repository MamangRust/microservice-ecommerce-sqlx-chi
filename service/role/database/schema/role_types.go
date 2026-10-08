// Package db — parameter and row structs kept from the former sqlc generation.
// Only the types still referenced by repositories/services live here; sqlx
// uses positional $n placeholders, so the params are plain data holders.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GetRolesParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}

type GetActiveRolesParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetActiveRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}

type GetTrashedRolesParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetTrashedRolesRow struct {
	RoleID     int32            `json:"role_id" db:"role_id"`
	RoleName   string           `json:"role_name" db:"role_name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	TotalCount int64            `json:"total_count" db:"total_count"`
}

type UpdateRoleParams struct {
	RoleID   int32  `json:"role_id" db:"role_id"`
	RoleName string `json:"role_name" db:"role_name"`
}

type AssignRoleToUserParams struct {
	UserID int32 `json:"user_id" db:"user_id"`
	RoleID int32 `json:"role_id" db:"role_id"`
}

type RemoveRoleFromUserParams struct {
	UserID int32 `json:"user_id" db:"user_id"`
	RoleID int32 `json:"role_id" db:"role_id"`
}
