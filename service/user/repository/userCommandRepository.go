package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/MamangRust/microservice-ecommerce-grpc-user/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/user_errors"
	"github.com/jmoiron/sqlx"
)

const createUser = `INSERT INTO
    users (
        firstname,
        lastname,
        email,
        password,
        verification_code,
        is_verified,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    )
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUser = `UPDATE users
SET
    firstname = $2,
    lastname = $3,
    email = $4,
    password = $5,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id, firstname, lastname, email, password, verification_code, is_verified, created_at, updated_at, deleted_at`

const updateUserIsVerified = `UPDATE users
SET
    is_verified = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const updateUserPassword = `UPDATE users
SET
    password = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at`

const trashUser = `UPDATE users
SET
    deleted_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const restoreUser = `UPDATE users
SET
    deleted_at = NULL
WHERE
    user_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    user_id,
    firstname,
    lastname,
    email,
    created_at,
    updated_at,
    deleted_at`

const deleteUserPermanently = `DELETE FROM users WHERE user_id = $1 AND deleted_at IS NOT NULL`

const restoreAllUsers = `UPDATE users
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentUsers = `DELETE FROM users WHERE deleted_at IS NOT NULL`

type userCommandRepository struct {
	db *sqlx.DB
}

func NewUserCommandRepository(db *sqlx.DB) *userCommandRepository {
	return &userCommandRepository{
		db: db,
	}
}

func (r *userCommandRepository) Create(ctx context.Context, request *requests.CreateUserRequest) (*db.CreateUserRow, error) {
	var user db.CreateUserRow
	err := r.db.GetContext(ctx, &user, createUser,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
		"",
		false,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrCreateUser.WithInternal(err)
	}

	return &user, nil
}

func (r *userCommandRepository) Update(ctx context.Context, request *requests.UpdateUserRequest) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, updateUser,
		int32(*request.UserID),
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrUpdateUser.WithInternal(err)
	}

	return &user, nil
}

func (r *userCommandRepository) Trash(ctx context.Context, user_id int) (*db.TrashUserRow, error) {
	var user db.TrashUserRow
	err := r.db.GetContext(ctx, &user, trashUser, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrTrashedUser.WithInternal(err)
	}

	return &user, nil
}

func (r *userCommandRepository) Restore(ctx context.Context, user_id int) (*db.RestoreUserRow, error) {
	var user db.RestoreUserRow
	err := r.db.GetContext(ctx, &user, restoreUser, int32(user_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrRestoreUser.WithInternal(err)
	}

	return &user, nil
}

func (r *userCommandRepository) DeletePermanent(ctx context.Context, user_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteUserPermanently, int32(user_id))

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete user while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, user_errors.ErrUserNotFound
		}
		return false, user_errors.ErrDeleteUserPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *userCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllUsers)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, user_errors.ErrUserNotFound
		}
		return false, user_errors.ErrRestoreAllUsers.WithInternal(err)
	}

	return true, nil
}

func (r *userCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentUsers)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete users while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, user_errors.ErrUserNotFound
		}
		return false, user_errors.ErrDeleteAllUsers.WithInternal(err)
	}
	return true, nil
}

func (r *userCommandRepository) UpdateIsVerified(ctx context.Context, user_id int, is_verified bool) (*db.UpdateUserIsVerifiedRow, error) {
	var user db.UpdateUserIsVerifiedRow
	err := r.db.GetContext(ctx, &user, updateUserIsVerified, int32(user_id), is_verified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrUpdateUser.WithInternal(err)
	}

	return &user, nil
}

func (r *userCommandRepository) UpdatePassword(ctx context.Context, user_id int, password string) (*db.UpdateUserPasswordRow, error) {
	var user db.UpdateUserPasswordRow
	err := r.db.GetContext(ctx, &user, updateUserPassword, int32(user_id), password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrUpdateUser.WithInternal(err)
	}

	return &user, nil
}
