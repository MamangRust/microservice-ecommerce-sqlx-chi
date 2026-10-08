package repository

import (
	"context"
	"database/sql"
	errorsstd "errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-banner/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/banner_errors"
	"github.com/jmoiron/sqlx"
)

const getBanners = `SELECT
    b.banner_id,
    b.name,
    b.start_date,
    b.end_date,
    b.start_time,
    b.end_time,
    b.is_active,
    b.created_at,
    b.updated_at,
    COUNT(*) OVER () AS total_count
FROM banners b
WHERE
    LOWER(name) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getBannersActive = `SELECT
    b.banner_id,
    b.name,
    b.start_date,
    b.end_date,
    b.start_time,
    b.end_time,
    b.is_active,
    b.created_at,
    b.updated_at,
    b.deleted_at,
    COUNT(*) OVER () AS total_count
FROM banners b
WHERE
    deleted_at IS NULL
    AND LOWER(name) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getBannersTrashed = `SELECT
    b.banner_id,
    b.name,
    b.start_date,
    b.end_date,
    b.start_time,
    b.end_time,
    b.is_active,
    b.created_at,
    b.updated_at,
    b.deleted_at,
    COUNT(*) OVER () AS total_count
FROM banners b
WHERE
    deleted_at IS NOT NULL
    AND LOWER(name) LIKE LOWER(CONCAT('%', $1::text, '%'))
LIMIT $2
OFFSET
    $3;`

const getBanner = `SELECT b.banner_id, b.name, b.start_date, b.end_date, b.start_time, b.end_time, b.is_active, b.created_at, b.updated_at
FROM banners b
WHERE
    banner_id = $1
    AND deleted_at IS NULL;`

type bannerQueryRepository struct {
	db *sqlx.DB
}

func NewBannerQueryRepository(db *sqlx.DB) *bannerQueryRepository {
	return &bannerQueryRepository{
		db: db,
	}
}

func (r *bannerQueryRepository) FindAll(ctx context.Context, req *requests.FindAllBanner) ([]*db.GetBannersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetBannersRow
	err := r.db.SelectContext(ctx, &res, getBanners,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, banner_errors.ErrFindAllBanners.WithInternal(err)
	}

	return res, nil
}

func (r *bannerQueryRepository) FindActive(ctx context.Context, req *requests.FindAllBanner) ([]*db.GetBannersActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetBannersActiveRow
	err := r.db.SelectContext(ctx, &res, getBannersActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, banner_errors.ErrFindActiveBanners.WithInternal(err)
	}

	return res, nil
}

func (r *bannerQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllBanner) ([]*db.GetBannersTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetBannersTrashedRow
	err := r.db.SelectContext(ctx, &res, getBannersTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, banner_errors.ErrFindTrashedBanners.WithInternal(err)
	}

	return res, nil
}

func (r *bannerQueryRepository) FindByID(ctx context.Context, bannerID int) (*db.GetBannerRow, error) {
	var res db.GetBannerRow
	err := r.db.GetContext(ctx, &res, getBanner, int32(bannerID))

	if err != nil {
		if errorsstd.Is(err, sql.ErrNoRows) {
			return nil, banner_errors.ErrBannerNotFound.WithInternal(err)
		}
		return nil, banner_errors.ErrFindAllBanners.WithInternal(err) // Generic find error
	}

	return &res, nil
}
