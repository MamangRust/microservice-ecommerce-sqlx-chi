package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-auth/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	refreshtoken_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/refresh_token_errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/refresh_token.sql).
const (
	createRefreshToken = `INSERT INTO refresh_tokens (user_id, token, expiration, created_at, updated_at)
VALUES ($1, $2, $3, current_timestamp, current_timestamp)
RETURNING refresh_token_id, user_id, token, expiration, created_at, updated_at, deleted_at`

	deleteRefreshToken = `DELETE FROM refresh_tokens
WHERE token = $1`

	deleteRefreshTokenByUserId = `DELETE FROM refresh_tokens
WHERE user_id = $1`

	findRefreshTokenByToken = `SELECT refresh_token_id, user_id, token, expiration, created_at, updated_at, deleted_at
FROM refresh_tokens
WHERE token = $1 AND deleted_at IS NULL`

	findRefreshTokenByUserId = `SELECT
    refresh_token_id,
    user_id,
    token,
    expiration,
    created_at,
    updated_at,
    deleted_at
FROM
    refresh_tokens
WHERE
    user_id = $1 AND deleted_at IS NULL
ORDER BY
    created_at DESC
LIMIT 1`

	updateRefreshTokenByUserId = `UPDATE refresh_tokens
SET token = $2, expiration = $3, updated_at = current_timestamp
WHERE user_id = $1 AND deleted_at IS NULL
RETURNING refresh_token_id, user_id, token, expiration, created_at, updated_at, deleted_at`
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *refreshTokenRepository {
	return &refreshTokenRepository{
		db: db,
	}
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*db.RefreshToken, error) {
	var res db.RefreshToken
	err := r.db.GetContext(ctx, &res, findRefreshTokenByToken, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, refreshtoken_errors.ErrTokenNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *refreshTokenRepository) FindByUserId(ctx context.Context, user_id int) (*db.RefreshToken, error) {
	var res db.RefreshToken
	err := r.db.GetContext(ctx, &res, findRefreshTokenByUserId, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, refreshtoken_errors.ErrTokenNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *refreshTokenRepository) CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*db.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, refreshtoken_errors.ErrParseDate
	}

	var res db.RefreshToken
	err = r.db.GetContext(ctx, &res, createRefreshToken,
		int32(req.UserId),
		req.Token,
		expirationTime,
	)
	if err != nil {
		return nil, refreshtoken_errors.ErrCreateRefreshToken.WithInternal(err)
	}

	return &res, nil
}

func (r *refreshTokenRepository) UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*db.RefreshToken, error) {
	layout := "2006-01-02 15:04:05"
	expirationTime, err := time.Parse(layout, req.ExpiresAt)
	if err != nil {
		return nil, refreshtoken_errors.ErrParseDate
	}

	var res db.RefreshToken
	err = r.db.GetContext(ctx, &res, updateRefreshTokenByUserId,
		int32(req.UserId),
		req.Token,
		expirationTime,
	)
	if err != nil {
		return nil, refreshtoken_errors.ErrUpdateRefreshToken.WithInternal(err)
	}

	return &res, nil
}

func (r *refreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, deleteRefreshToken, token)

	if err != nil {
		return refreshtoken_errors.ErrDeleteRefreshToken.WithInternal(err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error {
	_, err := r.db.ExecContext(ctx, deleteRefreshTokenByUserId, int32(user_id))

	if err != nil {
		return refreshtoken_errors.ErrDeleteByUserID.WithInternal(err)
	}

	return nil
}
