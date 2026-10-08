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

const createMerchantPolicy = `INSERT INTO
    merchant_policies (
        merchant_id,
        policy_type,
        title,
        description
    )
VALUES ($1, $2, $3, $4)
RETURNING
    merchant_policy_id,
    merchant_id,
    policy_type,
    title,
    description,
    created_at,
    updated_at`

const updateMerchantPolicy = `UPDATE merchant_policies
SET
    policy_type = $2,
    title = $3,
    description = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_policy_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_policy_id,
    merchant_id,
    policy_type,
    title,
    description,
    created_at,
    updated_at`

const trashMerchantPolicy = `UPDATE merchant_policies
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    merchant_policy_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_policy_id,
    merchant_id,
    policy_type,
    title,
    description,
    created_at,
    updated_at,
    deleted_at`

const restoreMerchantPolicy = `UPDATE merchant_policies
SET
    deleted_at = NULL
WHERE
    merchant_policy_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    merchant_policy_id,
    merchant_id,
    policy_type,
    title,
    description,
    created_at,
    updated_at,
    deleted_at`

const deleteMerchantPolicyPermanently = `DELETE FROM merchant_policies
WHERE
    merchant_policy_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllMerchantPolicies = `UPDATE merchant_policies
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllMerchantPolicyPermanently = `DELETE FROM merchant_policies WHERE deleted_at IS NOT NULL`

type merchantPolicyCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantPolicyCommandRepository(db *sqlx.DB) *merchantPolicyCommandRepository {
	return &merchantPolicyCommandRepository{
		db: db,
	}
}

func (r *merchantPolicyCommandRepository) Create(ctx context.Context, request *requests.CreateMerchantPolicyRequest) (*db.CreateMerchantPolicyRow, error) {
	var policy db.CreateMerchantPolicyRow

	if err := r.db.GetContext(ctx, &policy, createMerchantPolicy,
		int32(request.MerchantID),
		request.PolicyType,
		request.Title,
		request.Description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return nil, merchant_policy_errors.ErrCreateMerchantPolicy.WithInternal(err)
	}

	return &policy, nil
}

func (r *merchantPolicyCommandRepository) Update(ctx context.Context, request *requests.UpdateMerchantPolicyRequest) (*db.UpdateMerchantPolicyRow, error) {
	var res db.UpdateMerchantPolicyRow

	if err := r.db.GetContext(ctx, &res, updateMerchantPolicy,
		int32(*request.MerchantPolicyID),
		request.PolicyType,
		request.Title,
		request.Description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return nil, merchant_policy_errors.ErrUpdateMerchantPolicy.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantPolicyCommandRepository) Trash(ctx context.Context, merchant_policy_id int) (*db.MerchantPolicy, error) {
	var res db.MerchantPolicy

	if err := r.db.GetContext(ctx, &res, trashMerchantPolicy, int32(merchant_policy_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return nil, merchant_policy_errors.ErrTrashedMerchantPolicy.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantPolicyCommandRepository) Restore(ctx context.Context, merchant_policy_id int) (*db.MerchantPolicy, error) {
	var res db.MerchantPolicy

	if err := r.db.GetContext(ctx, &res, restoreMerchantPolicy, int32(merchant_policy_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return nil, merchant_policy_errors.ErrRestoreMerchantPolicy.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantPolicyCommandRepository) DeletePermanent(ctx context.Context, merchant_policy_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantPolicyPermanently, int32(merchant_policy_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return false, merchant_policy_errors.ErrDeleteMerchantPolicyPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *merchantPolicyCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchantPolicies); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return false, merchant_policy_errors.ErrRestoreAllMerchantPolicies.WithInternal(err)
	}
	return true, nil
}

func (r *merchantPolicyCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllMerchantPolicyPermanently); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_policy_errors.ErrMerchantPolicyNotFound
		}
		return false, merchant_policy_errors.ErrDeleteAllMerchantPoliciesPermanent.WithInternal(err)
	}
	return true, nil
}
