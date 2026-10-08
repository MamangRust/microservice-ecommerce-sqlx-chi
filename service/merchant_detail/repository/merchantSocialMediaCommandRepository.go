package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_detail/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchant_social_link_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant_social_link_errors"
	"github.com/jmoiron/sqlx"
)

const createMerchantSocialMediaLink = `INSERT INTO
    merchant_social_media_links (
        merchant_detail_id,
        platform,
        url
    )
VALUES ($1, $2, $3)
RETURNING
    merchant_social_id,
    merchant_detail_id,
    platform,
    url,
    created_at,
    updated_at`

const updateMerchantSocialMediaLink = `UPDATE merchant_social_media_links
SET
    platform = $2,
    url = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_social_id = $1
RETURNING
    merchant_social_id,
    merchant_detail_id,
    platform,
    url,
    created_at,
    updated_at`

const trashMerchantSocialMediaLink = `UPDATE merchant_social_media_links
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    merchant_social_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_social_id,
    merchant_detail_id,
    platform,
    url,
    created_at,
    updated_at,
    deleted_at`

const restoreMerchantSocialMediaLink = `UPDATE merchant_social_media_links
SET
    deleted_at = NULL
WHERE
    merchant_social_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    merchant_social_id,
    merchant_detail_id,
    platform,
    url,
    created_at,
    updated_at,
    deleted_at`

const deleteMerchantSocialMediaLinkPermanently = `DELETE FROM merchant_social_media_links
WHERE
    merchant_social_id = $1`

const restoreAllMerchantSocialMediaLinks = `UPDATE merchant_social_media_links
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllMerchantSocialMediaLinksPermanently = `DELETE FROM merchant_social_media_links WHERE deleted_at IS NOT NULL`

type merchantSocialLinkCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantSocialLinkCommandRepository(db *sqlx.DB) *merchantSocialLinkCommandRepository {
	return &merchantSocialLinkCommandRepository{
		db: db,
	}
}

func (r *merchantSocialLinkCommandRepository) Create(ctx context.Context, req *requests.CreateMerchantSocialRequest) (*db.CreateMerchantSocialMediaLinkRow, error) {
	var res db.CreateMerchantSocialMediaLinkRow

	if err := r.db.GetContext(ctx, &res, createMerchantSocialMediaLink,
		int32(*req.MerchantDetailID),
		req.Platform,
		req.Url,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return nil, merchant_social_link_errors.ErrCreateMerchantSocialLink.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantSocialLinkCommandRepository) Update(ctx context.Context, req *requests.UpdateMerchantSocialRequest) (*db.UpdateMerchantSocialMediaLinkRow, error) {
	var res db.UpdateMerchantSocialMediaLinkRow

	if err := r.db.GetContext(ctx, &res, updateMerchantSocialMediaLink,
		int32(req.ID),
		req.Platform,
		req.Url,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return nil, merchant_social_link_errors.ErrUpdateMerchantSocialLink.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantSocialLinkCommandRepository) Trash(ctx context.Context, socialID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, trashMerchantSocialMediaLink, int32(socialID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return false, merchant_social_link_errors.ErrTrashMerchantSocialLink.WithInternal(err)
	}

	return true, nil
}

func (r *merchantSocialLinkCommandRepository) Restore(ctx context.Context, socialID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreMerchantSocialMediaLink, int32(socialID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return false, merchant_social_link_errors.ErrRestoreMerchantSocialLink.WithInternal(err)
	}

	return true, nil
}

func (r *merchantSocialLinkCommandRepository) DeletePermanent(ctx context.Context, socialID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantSocialMediaLinkPermanently, int32(socialID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return false, merchant_social_link_errors.ErrDeletePermanentMerchantSocialLink.WithInternal(err)
	}

	return true, nil
}

func (r *merchantSocialLinkCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchantSocialMediaLinks); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return false, merchant_social_link_errors.ErrRestoreAllMerchantSocialLinks.WithInternal(err)
	}

	return true, nil
}

func (r *merchantSocialLinkCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllMerchantSocialMediaLinksPermanently); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchant_social_link_errors.ErrMerchantSocialLinkNotFound
		}
		return false, merchant_social_link_errors.ErrDeleteAllPermanentMerchantSocialLinks.WithInternal(err)
	}

	return true, nil
}
