package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type MerchantPolicy struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
