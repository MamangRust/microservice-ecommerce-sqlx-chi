package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchantdetail_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant_detail"
	"github.com/jmoiron/sqlx"
)

const createMerchantDetail = `INSERT INTO
    merchant_details (
        merchant_id,
        display_name,
        cover_image_url,
        logo_url,
        short_description,
        website_url
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    merchant_detail_id,
    merchant_id,
    display_name,
    cover_image_url,
    logo_url,
    short_description,
    website_url,
    created_at,
    updated_at`

const updateMerchantDetail = `UPDATE merchant_details
SET
    display_name = $2,
    cover_image_url = $3,
    logo_url = $4,
    short_description = $5,
    website_url = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_detail_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_detail_id,
    merchant_id,
    display_name,
    cover_image_url,
    logo_url,
    short_description,
    website_url,
    created_at,
    updated_at`

const trashMerchantDetail = `UPDATE merchant_details
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    merchant_detail_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_detail_id,
    merchant_id,
    display_name,
    cover_image_url,
    logo_url,
    short_description,
    website_url,
    created_at,
    updated_at,
    deleted_at`

const restoreMerchantDetail = `UPDATE merchant_details
SET
    deleted_at = NULL
WHERE
    merchant_detail_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    merchant_detail_id,
    merchant_id,
    display_name,
    cover_image_url,
    logo_url,
    short_description,
    website_url,
    created_at,
    updated_at,
    deleted_at`

const deleteMerchantDetailPermanently = `DELETE FROM merchant_details
WHERE
    merchant_detail_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllMerchantDetails = `UPDATE merchant_details
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentMerchantDetails = `DELETE FROM merchant_details WHERE deleted_at IS NOT NULL`

type merchantDetailCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantDetailCommandRepository(db *sqlx.DB) *merchantDetailCommandRepository {
	return &merchantDetailCommandRepository{
		db: db,
	}
}

func (r *merchantDetailCommandRepository) Create(ctx context.Context, request *requests.CreateMerchantDetailRequest) (*db.CreateMerchantDetailRow, error) {
	var merchant db.CreateMerchantDetailRow

	if err := r.db.GetContext(ctx, &merchant, createMerchantDetail,
		int32(request.MerchantID),
		&request.DisplayName,
		&request.CoverImageUrl,
		&request.LogoUrl,
		&request.ShortDescription,
		&request.WebsiteUrl,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return nil, merchantdetail_errors.ErrCreateMerchantDetail.WithInternal(err)
	}

	return &merchant, nil
}

func (r *merchantDetailCommandRepository) Update(ctx context.Context, request *requests.UpdateMerchantDetailRequest) (*db.UpdateMerchantDetailRow, error) {
	var res db.UpdateMerchantDetailRow

	if err := r.db.GetContext(ctx, &res, updateMerchantDetail,
		int32(*request.MerchantDetailID),
		&request.DisplayName,
		&request.CoverImageUrl,
		&request.LogoUrl,
		&request.ShortDescription,
		&request.WebsiteUrl,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return nil, merchantdetail_errors.ErrUpdateMerchantDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantDetailCommandRepository) Trash(ctx context.Context, merchant_detail_id int) (*db.MerchantDetail, error) {
	var res db.MerchantDetail

	if err := r.db.GetContext(ctx, &res, trashMerchantDetail, int32(merchant_detail_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return nil, merchantdetail_errors.ErrTrashMerchantDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantDetailCommandRepository) Restore(ctx context.Context, merchant_detail_id int) (*db.MerchantDetail, error) {
	var res db.MerchantDetail

	if err := r.db.GetContext(ctx, &res, restoreMerchantDetail, int32(merchant_detail_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return nil, merchantdetail_errors.ErrRestoreMerchantDetail.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantDetailCommandRepository) DeletePermanent(ctx context.Context, merchant_detail_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantDetailPermanently, int32(merchant_detail_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return false, merchantdetail_errors.ErrDeletePermanentMerchantDetail.WithInternal(err)
	}

	return true, nil
}

func (r *merchantDetailCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchantDetails); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return false, merchantdetail_errors.ErrRestoreAllMerchantDetails.WithInternal(err)
	}
	return true, nil
}

func (r *merchantDetailCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentMerchantDetails); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantdetail_errors.ErrMerchantDetailNotFound
		}
		return false, merchantdetail_errors.ErrDeleteAllPermanentMerchantDetails.WithInternal(err)
	}
	return true, nil
}
