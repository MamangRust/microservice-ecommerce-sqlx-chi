package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-role/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/role_errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/roles.sql).
const (
	createRole = `INSERT INTO roles (
    role_name,
    created_at,
    updated_at
) VALUES (
    $1,
    current_timestamp,
    current_timestamp
) RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`

	deleteAllPermanentRoles = `DELETE FROM roles
WHERE
    deleted_at IS NOT NULL`

	deletePermanentRole = `DELETE FROM roles
WHERE
    role_id = $1 AND deleted_at IS NOT NULL`

	restoreAllRoles = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

	restoreRole = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    role_id = $1
RETURNING role_id, role_name, created_at, updated_at, deleted_at`

	trashRole = `UPDATE roles
SET
    deleted_at = current_timestamp
WHERE
    role_id = $1
RETURNING role_id, role_name, created_at, updated_at, deleted_at`

	updateRole = `UPDATE roles
SET
    role_name = $2,
    updated_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`
)

type roleCommandRepository struct {
	db *sqlx.DB
}

func NewRoleCommandRepository(db *sqlx.DB) *roleCommandRepository {
	return &roleCommandRepository{
		db: db,
	}
}

func (r *roleCommandRepository) Create(ctx context.Context, req *requests.CreateRoleRequest) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, createRole, req.Name)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, role_errors.ErrRoleConflict
		}
		return nil, role_errors.ErrCreateRole.WithInternal(err)
	}

	return &res, nil
}

func (r *roleCommandRepository) Update(ctx context.Context, req *requests.UpdateRoleRequest) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, updateRole,
		int32(*req.ID),
		req.Name,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, role_errors.ErrRoleConflict
		}
		return nil, role_errors.ErrUpdateRole.WithInternal(err)
	}

	return &res, nil
}

func (r *roleCommandRepository) Trash(ctx context.Context, id int) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, trashRole, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, role_errors.ErrRoleNotFound
		}
		return nil, role_errors.ErrTrashedRole.WithInternal(err)
	}
	return &res, nil
}

func (r *roleCommandRepository) Restore(ctx context.Context, id int) (*db.Role, error) {
	var res db.Role
	err := r.db.GetContext(ctx, &res, restoreRole, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, role_errors.ErrRoleNotFound
		}
		return nil, role_errors.ErrRestoreRole.WithInternal(err)
	}
	return &res, nil
}

func (r *roleCommandRepository) DeletePermanent(ctx context.Context, role_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deletePermanentRole, int32(role_id))
	if err != nil {
		return false, role_errors.ErrDeleteRolePermanent.WithInternal(err)
	}
	return true, nil
}

func (r *roleCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllRoles)

	if err != nil {
		return false, role_errors.ErrRestoreAllRoles.WithInternal(err)
	}

	return true, nil
}

func (r *roleCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentRoles)

	if err != nil {
		return false, role_errors.ErrDeleteAllRoles.WithInternal(err)
	}

	return true, nil
}
