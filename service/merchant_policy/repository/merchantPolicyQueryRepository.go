package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchant_policy_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant_policy_errors"
	"github.com/jmoiron/sqlx"
)

const getMerchantPolicies = `SELECT
    mp.merchant_policy_id,
    mp.merchant_id,
    mp.policy_type,
    mp.title,
    mp.description,
    mp.created_at,
    mp.updated_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count
FROM
    merchant_policies mp
WHERE
    mp.deleted_at IS NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
LIMIT $2
OFFSET
    $3`

const getMerchantPoliciesActive = `SELECT
    mp.merchant_policy_id,
    mp.merchant_id,
    mp.policy_type,
    mp.title,
    mp.description,
    mp.created_at,
    mp.updated_at,
    mp.deleted_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count
FROM
    merchant_policies mp
WHERE
    mp.deleted_at IS NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
LIMIT $2
OFFSET
    $3`

const getMerchantPoliciesTrashed = `SELECT
    mp.merchant_policy_id,
    mp.merchant_id,
    mp.policy_type,
    mp.title,
    mp.description,
    mp.created_at,
    mp.updated_at,
    mp.deleted_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count
FROM
    merchant_policies mp
WHERE
    mp.deleted_at IS NOT NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
LIMIT $2
OFFSET
    $3`

const getMerchantPolicy = `SELECT mp.merchant_policy_id, mp.merchant_id, mp.policy_type, mp.title, mp.description, mp.created_at, mp.updated_at
FROM merchant_policies mp
WHERE
    merchant_policy_id = $1
    AND deleted_at IS NULL`

type merchantPolicyQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantPolicyQueryRepository(db *sqlx.DB) *merchantPolicyQueryRepository {
	return &merchantPolicyQueryRepository{
		db: db,
	}
}

func (r *merchantPolicyQueryRepository) FindAll(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantPoliciesRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantPoliciesRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantPolicies, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchant_policy_errors.ErrFindAllMerchantPolicies.WithInternal(err)
	}

	return res, nil
}

func (r *merchantPolicyQueryRepository) FindActive(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantPoliciesActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantPoliciesActiveRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantPoliciesActive, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchant_policy_errors.ErrFindActiveMerchantPolicies.WithInternal(err)
	}

	return res, nil
}

func (r *merchantPolicyQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantPoliciesTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantPoliciesTrashedRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantPoliciesTrashed, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchant_policy_errors.ErrFindTrashedMerchantPolicies.WithInternal(err)
	}

	return res, nil
}

func (r *merchantPolicyQueryRepository) FindByID(ctx context.Context, user_id int) (*db.GetMerchantPolicyRow, error) {
	var res db.GetMerchantPolicyRow

	if err := r.db.GetContext(ctx, &res, getMerchantPolicy, int32(user_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return nil, merchant_policy_errors.ErrFindMerchantPolicyByID.WithInternal(err)
	}

	return &res, nil
}
