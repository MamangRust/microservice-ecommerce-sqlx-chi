package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-user/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/user_errors"
	"github.com/jmoiron/sqlx"
)

const getUsers = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getUsersActive = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getUserTrashed = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM users
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR firstname ILIKE '%' || $1 || '%'
        OR lastname ILIKE '%' || $1 || '%'
        OR email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getUserByID = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    user_id = $1
    AND deleted_at IS NULL`

const getUserByEmail = `SELECT user_id, firstname, lastname, email, password, verification_code, is_verified, created_at, updated_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL`

const getUserByEmailWithPassword = `SELECT user_id, email, password
FROM users
WHERE
    email = $1
    AND deleted_at IS NULL`

const getUserByVerificationCode = `SELECT
    user_id,
    firstname,
    lastname,
    email,
    password,
    created_at,
    updated_at
FROM users
WHERE
    verification_code = $1`

type userQueryRepository struct {
	db *sqlx.DB
}

func NewUserQueryRepository(db *sqlx.DB) *userQueryRepository {
	return &userQueryRepository{
		db: db,
	}
}

func (r *userQueryRepository) FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUsersRow
	err := r.db.SelectContext(ctx, &res, getUsers,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, user_errors.ErrFindAllUsers.WithInternal(err)
	}

	return res, nil
}

func (r *userQueryRepository) FindByID(ctx context.Context, user_id int) (*db.GetUserByIDRow, error) {
	var user db.GetUserByIDRow
	err := r.db.GetContext(ctx, &user, getUserByID, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByIDWithPassword(ctx context.Context, user_id int) (*db.GetUserByIDRow, error) {
	var user db.GetUserByIDRow
	err := r.db.GetContext(ctx, &user, getUserByID, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindActive(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUsersActiveRow
	err := r.db.SelectContext(ctx, &res, getUsersActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, user_errors.ErrFindActiveUsers.WithInternal(err)
	}

	return res, nil
}

func (r *userQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUserTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUserTrashedRow
	err := r.db.SelectContext(ctx, &res, getUserTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, user_errors.ErrFindTrashedUsers.WithInternal(err)
	}

	return res, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByEmailWithPassword(ctx context.Context, email string) (*db.GetUserByEmailWithPasswordRow, error) {
	var user db.GetUserByEmailWithPasswordRow
	err := r.db.GetContext(ctx, &user, getUserByEmailWithPassword, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByVerificationCode(ctx context.Context, code string) (*db.GetUserByVerificationCodeRow, error) {
	var user db.GetUserByVerificationCodeRow
	err := r.db.GetContext(ctx, &user, getUserByVerificationCode, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &user, nil
}
