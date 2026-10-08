// Package db holds the database row models for the auth service.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxEvent struct {
	OutboxID      int64            `json:"outbox_id" db:"outbox_id"`
	Topic         string           `json:"topic" db:"topic"`
	EventKey      string           `json:"event_key" db:"event_key"`
	Payload       []byte           `json:"payload" db:"payload"`
	Status        string           `json:"status" db:"status"`
	Attempts      int32            `json:"attempts" db:"attempts"`
	NextAttemptAt time.Time        `json:"next_attempt_at" db:"next_attempt_at"`
	CreatedAt     pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt     pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type RefreshToken struct {
	RefreshTokenID int32            `json:"refresh_token_id" db:"refresh_token_id"`
	UserID         int32            `json:"user_id" db:"user_id"`
	Token          string           `json:"token" db:"token"`
	Expiration     time.Time        `json:"expiration" db:"expiration"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type ResetToken struct {
	ID         int32     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}
