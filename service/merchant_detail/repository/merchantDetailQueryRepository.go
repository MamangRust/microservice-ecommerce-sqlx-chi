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

const getMerchantDetails = `SELECT
    md.merchant_detail_id,
    md.merchant_id,
    md.display_name,
    md.cover_image_url,
    md.logo_url,
    md.short_description,
    md.website_url,
    md.created_at,
    md.updated_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count,
    json_agg(
        json_build_object(
            'id',
            sml.merchant_social_id,
            'platform',
            sml.platform,
            'url',
            sml.url
        )
    ) AS social_media_links
FROM
    merchant_details md
    LEFT JOIN merchant_social_media_links sml ON sml.merchant_detail_id = md.merchant_detail_id
WHERE
    md.deleted_at IS NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
GROUP BY
    md.merchant_detail_id
LIMIT $2
OFFSET
    $3`

const getMerchantDetailsActive = `SELECT
    md.merchant_detail_id,
    md.merchant_id,
    md.display_name,
    md.cover_image_url,
    md.logo_url,
    md.short_description,
    md.website_url,
    md.created_at,
    md.updated_at,
    md.deleted_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count,
    json_agg(
        json_build_object(
            'id',
            sml.merchant_social_id,
            'platform',
            sml.platform,
            'url',
            sml.url
        )
    ) AS social_media_links
FROM
    merchant_details md
    LEFT JOIN merchant_social_media_links sml ON sml.merchant_detail_id = md.merchant_detail_id
WHERE
    md.deleted_at IS NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
GROUP BY
    md.merchant_detail_id
LIMIT $2
OFFSET
    $3`

const getMerchantDetailsTrashed = `SELECT
    md.merchant_detail_id,
    md.merchant_id,
    md.display_name,
    md.cover_image_url,
    md.logo_url,
    md.short_description,
    md.website_url,
    md.created_at,
    md.updated_at,
    md.deleted_at,
    ''::text AS merchant_name,
    COUNT(*) OVER () AS total_count,
    json_agg(
        json_build_object(
            'id',
            sml.merchant_social_id,
            'platform',
            sml.platform,
            'url',
            sml.url
        )
    ) AS social_media_links
FROM
    merchant_details md
    LEFT JOIN merchant_social_media_links sml ON sml.merchant_detail_id = md.merchant_detail_id
WHERE
    md.deleted_at IS NOT NULL
    AND ($1::TEXT IS NULL OR $1::TEXT = '')
GROUP BY
    md.merchant_detail_id
LIMIT $2
OFFSET
    $3`

const getMerchantDetail = `SELECT
    md.merchant_detail_id,
    md.merchant_id,
    md.display_name,
    md.cover_image_url,
    md.logo_url,
    md.short_description,
    md.website_url,
    md.created_at,
    md.updated_at,
    ''::text AS merchant_name,
    json_agg(
        json_build_object(
            'id',
            sml.merchant_social_id,
            'platform',
            sml.platform,
            'url',
            sml.url
        )
    ) AS social_media_links
FROM
    merchant_details md
    LEFT JOIN merchant_social_media_links sml ON sml.merchant_detail_id = md.merchant_detail_id
WHERE
    md.merchant_detail_id = $1
    AND md.deleted_at IS NULL
GROUP BY
    md.merchant_detail_id`

const getMerchantDetailTrashed = `SELECT md.merchant_detail_id, md.merchant_id, md.display_name, md.cover_image_url, md.logo_url, md.short_description, md.website_url, md.created_at, md.updated_at
FROM merchant_details md
WHERE
    merchant_detail_id = $1
    AND deleted_at IS NOT NULL`

type merchantDetailQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantDetailQueryRepository(db *sqlx.DB) *merchantDetailQueryRepository {
	return &merchantDetailQueryRepository{
		db: db,
	}
}

func (r *merchantDetailQueryRepository) FindAll(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantDetailsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantDetailsRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantDetails, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchantdetail_errors.ErrFindAllMerchantDetails.WithInternal(err)
	}

	return res, nil
}

func (r *merchantDetailQueryRepository) FindActive(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantDetailsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantDetailsActiveRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantDetailsActive, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchantdetail_errors.ErrFindActiveMerchantDetails.WithInternal(err)
	}

	return res, nil
}

func (r *merchantDetailQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllMerchant) ([]*db.GetMerchantDetailsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	res := []*db.GetMerchantDetailsTrashedRow{}

	if err := r.db.SelectContext(ctx, &res, getMerchantDetailsTrashed, req.Search, int32(req.PageSize), int32(offset)); err != nil {
		return nil, merchantdetail_errors.ErrFindTrashedMerchantDetails.WithInternal(err)
	}

	return res, nil
}

func (r *merchantDetailQueryRepository) FindByID(ctx context.Context, user_id int) (*db.GetMerchantDetailRow, error) {
	var res db.GetMerchantDetailRow

	if err := r.db.GetContext(ctx, &res, getMerchantDetail, int32(user_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound.WithInternal(err)
		}
		return nil, merchantdetail_errors.ErrMerchantDetailInternal.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantDetailQueryRepository) FindByIDTrashed(ctx context.Context, user_id int) (*db.GetMerchantDetailTrashedRow, error) {
	var res db.GetMerchantDetailTrashedRow

	if err := r.db.GetContext(ctx, &res, getMerchantDetailTrashed, int32(user_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantdetail_errors.ErrMerchantDetailNotFound.WithInternal(err)
		}
		return nil, merchantdetail_errors.ErrFindByIdTrashedMerchantDetail.WithInternal(err)
	}

	return &res, nil
}
