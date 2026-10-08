// Code migrated from sqlc-generated sources (reviews.sql.go); query row structs
// kept for the repository layer with sqlx db tags added.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GetReviewsRow struct {
	ReviewID   int32            `db:"review_id" json:"review_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	ProductID  int32            `db:"product_id" json:"product_id"`
	Name       string           `db:"name" json:"name"`
	Comment    string           `db:"comment" json:"comment"`
	Rating     int32            `db:"rating" json:"rating"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetReviewsActiveRow struct {
	ReviewID   int32            `db:"review_id" json:"review_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	ProductID  int32            `db:"product_id" json:"product_id"`
	Name       string           `db:"name" json:"name"`
	Comment    string           `db:"comment" json:"comment"`
	Rating     int32            `db:"rating" json:"rating"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetReviewsTrashedRow struct {
	ReviewID   int32            `db:"review_id" json:"review_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	ProductID  int32            `db:"product_id" json:"product_id"`
	Name       string           `db:"name" json:"name"`
	Comment    string           `db:"comment" json:"comment"`
	Rating     int32            `db:"rating" json:"rating"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetReviewByProductIdRow struct {
	ReviewID      int32            `db:"review_id" json:"review_id"`
	UserID        int32            `db:"user_id" json:"user_id"`
	ProductID     int32            `db:"product_id" json:"product_id"`
	Name          string           `db:"name" json:"name"`
	Comment       string           `db:"comment" json:"comment"`
	Rating        int32            `db:"rating" json:"rating"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
	ReviewDetails []byte           `db:"review_details" json:"review_details"`
}

type GetReviewByMerchantIdRow struct {
	ReviewID      int32            `db:"review_id" json:"review_id"`
	UserID        int32            `db:"user_id" json:"user_id"`
	ProductID     int32            `db:"product_id" json:"product_id"`
	Name          string           `db:"name" json:"name"`
	Comment       string           `db:"comment" json:"comment"`
	Rating        int32            `db:"rating" json:"rating"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount    int64            `db:"total_count" json:"total_count"`
	ReviewDetails []byte           `db:"review_details" json:"review_details"`
}

type GetReviewByIDRow struct {
	ReviewID  int32            `db:"review_id" json:"review_id"`
	UserID    int32            `db:"user_id" json:"user_id"`
	ProductID int32            `db:"product_id" json:"product_id"`
	Name      string           `db:"name" json:"name"`
	Comment   string           `db:"comment" json:"comment"`
	Rating    int32            `db:"rating" json:"rating"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type CreateReviewRow struct {
	ReviewID  int32            `db:"review_id" json:"review_id"`
	UserID    int32            `db:"user_id" json:"user_id"`
	ProductID int32            `db:"product_id" json:"product_id"`
	Name      string           `db:"name" json:"name"`
	Comment   string           `db:"comment" json:"comment"`
	Rating    int32            `db:"rating" json:"rating"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type UpdateReviewRow struct {
	ReviewID  int32            `db:"review_id" json:"review_id"`
	UserID    int32            `db:"user_id" json:"user_id"`
	ProductID int32            `db:"product_id" json:"product_id"`
	Name      string           `db:"name" json:"name"`
	Comment   string           `db:"comment" json:"comment"`
	Rating    int32            `db:"rating" json:"rating"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
