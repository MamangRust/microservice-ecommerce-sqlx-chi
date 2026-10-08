package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserParams struct {
	Firstname        string `db:"firstname" json:"firstname"`
	Lastname         string `db:"lastname" json:"lastname"`
	Email            string `db:"email" json:"email"`
	Password         string `db:"password" json:"password"`
	VerificationCode string `db:"verification_code" json:"verification_code"`
	IsVerified       *bool  `db:"is_verified" json:"is_verified"`
}

type CreateUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUserByIDRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUserByEmailWithPasswordRow struct {
	UserID   int32  `db:"user_id" json:"user_id"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}

type GetUserByVerificationCodeRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type GetUsersRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetUsersActiveRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetUserTrashedRow struct {
	UserID     int32            `db:"user_id" json:"user_id"`
	Firstname  string           `db:"firstname" json:"firstname"`
	Lastname   string           `db:"lastname" json:"lastname"`
	Email      string           `db:"email" json:"email"`
	Password   string           `db:"password" json:"password"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type TrashUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type RestoreUserRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type UpdateUserIsVerifiedParams struct {
	UserID     int32 `db:"user_id" json:"user_id"`
	IsVerified *bool `db:"is_verified" json:"is_verified"`
}

type UpdateUserIsVerifiedRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateUserPasswordParams struct {
	UserID   int32  `db:"user_id" json:"user_id"`
	Password string `db:"password" json:"password"`
}

type UpdateUserPasswordRow struct {
	UserID    int32            `db:"user_id" json:"user_id"`
	Firstname string           `db:"firstname" json:"firstname"`
	Lastname  string           `db:"lastname" json:"lastname"`
	Email     string           `db:"email" json:"email"`
	Password  string           `db:"password" json:"password"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
