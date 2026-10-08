package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	UserID           int32            `db:"user_id" json:"user_id"`
	Firstname        string           `db:"firstname" json:"firstname"`
	Lastname         string           `db:"lastname" json:"lastname"`
	Email            string           `db:"email" json:"email"`
	Password         string           `db:"password" json:"password"`
	VerificationCode string           `db:"verification_code" json:"verification_code"`
	IsVerified       *bool            `db:"is_verified" json:"is_verified"`
	CreatedAt        pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt        pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt        pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
