package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-auth/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	resettoken_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/reset_token_errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/reset_token.sql).
const (
	createResetToken = `INSERT INTO
    reset_tokens (user_id, token, expiry_date)
VALUES ($1, $2, $3)
RETURNING
    id, user_id, token, expiry_date`

	deleteResetToken = `DELETE FROM reset_tokens WHERE user_id = $1`

	getResetToken = `SELECT id, user_id, token, expiry_date FROM reset_tokens WHERE token = $1`
)

// resetTokenRepository is a struct that implements the ResetTokenRepository interface
type resetTokenRepository struct {
	db *sqlx.DB
}

// NewResetTokenRepository creates a new instance of resetTokenRepository.
func NewResetTokenRepository(db *sqlx.DB) *resetTokenRepository {
	return &resetTokenRepository{
		db: db,
	}
}

// FindByToken retrieves a reset token record by token string.
func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*db.ResetToken, error) {
	var res db.ResetToken
	err := r.db.GetContext(ctx, &res, getResetToken, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, resettoken_errors.ErrTokenNotFound.WithInternal(err)
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &res, nil
}

// CreateResetTokenInTx persists the reset token inside the given database
// transaction so the caller can commit the token write and its outbox event
// atomically (Phase 6 — transactional outbox).
func (r *resetTokenRepository) CreateResetTokenInTx(ctx context.Context, tx *sqlx.Tx, req *requests.CreateResetTokenRequest) (*db.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, err
	}
	var res db.ResetToken
	err = tx.GetContext(ctx, &res, createResetToken,
		int64(req.UserID),
		req.ResetToken,
		expiryDate,
	)
	if err != nil {
		return nil, resettoken_errors.ErrCreateResetToken.WithInternal(err)
	}
	return &res, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, err
	}
	var res db.ResetToken
	err = r.db.GetContext(ctx, &res, createResetToken,
		int64(req.UserID),
		req.ResetToken,
		expiryDate,
	)
	if err != nil {
		return nil, resettoken_errors.ErrCreateResetToken.WithInternal(err)
	}

	return &res, nil
}

// DeleteResetToken removes the reset token associated with the given user ID.
func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, user_id int) error {
	_, err := r.db.ExecContext(ctx, deleteResetToken, int64(user_id))
	if err != nil {
		return resettoken_errors.ErrDeleteByUserID.WithInternal(err)
	}

	return nil
}
