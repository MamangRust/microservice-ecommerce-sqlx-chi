// Package db — parameter structs kept from the former sqlc generation.
// Only the types still referenced by repositories/services live here; sqlx
// uses positional $n placeholders, so these are plain data holders.
package db

import (
	"time"
)

type CreateRefreshTokenParams struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Expiration time.Time `json:"expiration" db:"expiration"`
}

type UpdateRefreshTokenByUserIdParams struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Expiration time.Time `json:"expiration" db:"expiration"`
}

type CreateResetTokenParams struct {
	UserID     int64     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}
