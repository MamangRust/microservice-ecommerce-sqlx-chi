package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
)

type merchantCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantCommandRepository(db *sqlx.DB) MerchantCommandRepository {
	return &merchantCommandRepository{
		db: db,
	}
}

const createMerchant = `INSERT INTO
    merchants (
        user_id,
        name,
        description,
        address,
        contact_email,
        contact_phone,
        status
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at`

const updateMerchant = `UPDATE merchants
SET
    name = $2,
    description = $3,
    address = $4,
    contact_email = $5,
    contact_phone = $6,
    status = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at`

const updateMerchantStatus = `UPDATE merchants
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at`

const trashMerchant = `UPDATE merchants
SET
    deleted_at = current_timestamp
WHERE
    merchant_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    deleted_at`

const restoreMerchant = `UPDATE merchants
SET
    deleted_at = NULL
WHERE
    merchant_id = $1
    AND deleted_at IS NOT NULL
RETURNING
     merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    deleted_at`

const deleteMerchantPermanently = `DELETE FROM merchants
WHERE
    merchant_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllMerchants = `UPDATE merchants
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentMerchants = `DELETE FROM merchants WHERE deleted_at IS NOT NULL`

func (r *merchantCommandRepository) Create(
	ctx context.Context,
	request *requests.CreateMerchantRequest,
) (*db.CreateMerchantRow, error) {
	return r.createMerchant(ctx, r.db, request)
}

// CreateInTx persists the merchant inside the given database transaction so the
// caller can commit the business write and its outbox event atomically (Phase 6
// — transactional outbox).
func (r *merchantCommandRepository) CreateInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	request *requests.CreateMerchantRequest,
) (*db.CreateMerchantRow, error) {
	return r.createMerchant(ctx, tx, request)
}

func (r *merchantCommandRepository) createMerchant(ctx context.Context, q sqlx.QueryerContext, request *requests.CreateMerchantRequest) (*db.CreateMerchantRow, error) {
	var merchant db.CreateMerchantRow
	err := sqlx.GetContext(ctx, q, &merchant, createMerchant,
		int32(request.UserID),
		request.Name,
		stringPtr(request.Description),
		stringPtr(request.Address),
		stringPtr(request.ContactEmail),
		stringPtr(request.ContactPhone),
		"active",
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound
		}
		return nil, merchant_errors.ErrCreateMerchant.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) Update(ctx context.Context, request *requests.UpdateMerchantRequest) (*db.UpdateMerchantRow, error) {
	var merchant db.UpdateMerchantRow
	err := r.db.GetContext(ctx, &merchant, updateMerchant,
		int32(*request.MerchantID),
		request.Name,
		stringPtr(request.Description),
		stringPtr(request.Address),
		stringPtr(request.ContactEmail),
		stringPtr(request.ContactPhone),
		request.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound
		}
		return nil, merchant_errors.ErrUpdateMerchant.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) Trash(ctx context.Context, merchant_id int) (*db.Merchant, error) {
	var merchant db.Merchant
	err := r.db.GetContext(ctx, &merchant, trashMerchant, int32(merchant_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound
		}
		return nil, merchant_errors.ErrTrashedMerchant.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) Restore(ctx context.Context, merchant_id int) (*db.Merchant, error) {
	var merchant db.Merchant
	err := r.db.GetContext(ctx, &merchant, restoreMerchant, int32(merchant_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound
		}
		return nil, merchant_errors.ErrRestoreMerchant.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantCommandRepository) DeletePermanent(ctx context.Context, Merchant_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteMerchantPermanently, int32(Merchant_id))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete merchant while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_errors.ErrMerchantNotFound
		}
		return false, merchant_errors.ErrDeleteMerchantPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *merchantCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllMerchants)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_errors.ErrMerchantNotFound
		}
		return false, merchant_errors.ErrRestoreAllMerchants.WithInternal(err)
	}

	return true, nil
}

func (r *merchantCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentMerchants)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete merchants while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_errors.ErrMerchantNotFound
		}
		return false, merchant_errors.ErrDeleteAllMerchants.WithInternal(err)
	}

	return true, nil
}

func (r *merchantCommandRepository) UpdateStatus(ctx context.Context, request *requests.UpdateMerchantStatusRequest) (*db.UpdateMerchantStatusRow, error) {
	return r.updateMerchantStatus(ctx, r.db, request)
}

// UpdateStatusInTx updates the merchant status inside the given database
// transaction so the caller can commit the business write and its outbox event
// atomically (Phase 6 — transactional outbox).
func (r *merchantCommandRepository) UpdateStatusInTx(ctx context.Context, tx *sqlx.Tx, request *requests.UpdateMerchantStatusRequest) (*db.UpdateMerchantStatusRow, error) {
	return r.updateMerchantStatus(ctx, tx, request)
}

func (r *merchantCommandRepository) updateMerchantStatus(ctx context.Context, q sqlx.QueryerContext, request *requests.UpdateMerchantStatusRequest) (*db.UpdateMerchantStatusRow, error) {
	var merchant db.UpdateMerchantStatusRow
	err := sqlx.GetContext(ctx, q, &merchant, updateMerchantStatus,
		int32(*request.MerchantID),
		request.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound
		}
		return nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	return &merchant, nil
}
