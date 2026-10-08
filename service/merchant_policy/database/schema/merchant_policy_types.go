package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateMerchantPolicyRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateMerchantPolicyRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantPolicyRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantPoliciesRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
}

type GetMerchantPoliciesActiveRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
}

type GetMerchantPoliciesTrashedRow struct {
	MerchantPolicyID int32            `json:"merchant_policy_id" db:"merchant_policy_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	PolicyType       string           `json:"policy_type" db:"policy_type"`
	Title            string           `json:"title" db:"title"`
	Description      string           `json:"description" db:"description"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
}
