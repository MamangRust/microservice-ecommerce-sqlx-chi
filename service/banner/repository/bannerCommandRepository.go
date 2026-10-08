package repository

import (
	"context"
	"database/sql"
	errorsstd "errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-banner/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/utils"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/banner_errors"
	"github.com/jmoiron/sqlx"
)

const createBanner = `INSERT INTO
    banners (
        name,
        start_date,
        end_date,
        start_time,
        end_time,
        is_active
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    banner_id,
    name,
    start_date,
    end_date,
    start_time,
    end_time,
    is_active,
    created_at,
    updated_at;`

const updateBanner = `UPDATE banners
SET
    name = $2,
    start_date = $3,
    end_date = $4,
    start_time = $5,
    end_time = $6,
    is_active = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE
    banner_id = $1
    AND deleted_at IS NULL
RETURNING
    banner_id,
    name,
    start_date,
    end_date,
    start_time,
    end_time,
    is_active,
    created_at,
    updated_at;`

const trashBanner = `UPDATE banners
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    banner_id = $1
    AND deleted_at IS NULL
RETURNING
    banner_id,
    name,
    start_date,
    end_date,
    start_time,
    end_time,
    is_active,
    created_at,
    updated_at,
    deleted_at;`

const restoreBanner = `UPDATE banners
SET
    deleted_at = NULL
WHERE
    banner_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    banner_id,
    name,
    start_date,
    end_date,
    start_time,
    end_time,
    is_active,
    created_at,
    updated_at,
    deleted_at;`

const deleteBannerPermanently = `DELETE FROM banners
WHERE
    banner_id = $1
    AND deleted_at IS NOT NULL;`

const restoreAllBanners = `UPDATE banners SET deleted_at = NULL WHERE deleted_at IS NOT NULL;`

const deleteAllPermanentBanners = `DELETE FROM banners WHERE deleted_at IS NOT NULL;`

type bannerCommandRepository struct {
	db *sqlx.DB
}

func NewBannerCommandRepository(db *sqlx.DB) *bannerCommandRepository {
	return &bannerCommandRepository{
		db: db,
	}
}

func (r *bannerCommandRepository) Create(ctx context.Context, request *requests.CreateBannerRequest) (*db.CreateBannerRow, error) {
	startDate, err := utils.ParseDate(request.StartDate)
	if err != nil {
		return nil, banner_errors.ErrBannerStartDate
	}

	endDate, err := utils.ParseDate(request.EndDate)
	if err != nil {
		return nil, banner_errors.ErrBannerEndDate
	}

	startTime, err := utils.ParseTime(request.StartTime)
	if err != nil {
		return nil, banner_errors.ErrBannerStartTime
	}

	endTime, err := utils.ParseTime(request.EndTime)
	if err != nil {
		return nil, banner_errors.ErrBannerEndTime
	}

	var result db.CreateBannerRow
	err = r.db.GetContext(ctx, &result, createBanner,
		request.Name,
		startDate,
		endDate,
		startTime,
		endTime,
		&request.IsActive,
	)
	if err != nil {
		return nil, banner_errors.ErrCreateBanner.WithInternal(err)
	}

	return &result, nil
}

func (r *bannerCommandRepository) Update(ctx context.Context, request *requests.UpdateBannerRequest) (*db.UpdateBannerRow, error) {
	startDate, err := utils.ParseDate(request.StartDate)
	if err != nil {
		return nil, banner_errors.ErrBannerStartDate.WithInternal(err)
	}

	endDate, err := utils.ParseDate(request.EndDate)
	if err != nil {
		return nil, banner_errors.ErrBannerEndDate.WithInternal(err)
	}

	startTime, err := utils.ParseTime(request.StartTime)
	if err != nil {
		return nil, banner_errors.ErrBannerStartTime.WithInternal(err)
	}

	endTime, err := utils.ParseTime(request.EndTime)
	if err != nil {
		return nil, banner_errors.ErrBannerEndTime.WithInternal(err)
	}

	var result db.UpdateBannerRow
	err = r.db.GetContext(ctx, &result, updateBanner,
		int32(*request.BannerID),
		request.Name,
		startDate,
		endDate,
		startTime,
		endTime,
		&request.IsActive,
	)
	if err != nil {
		return nil, banner_errors.ErrUpdateBanner.WithInternal(err)
	}

	return &result, nil
}

func (r *bannerCommandRepository) Trash(ctx context.Context, bannerID int) (*db.Banner, error) {
	var res db.Banner
	err := r.db.GetContext(ctx, &res, trashBanner, int32(bannerID))

	if err != nil {
		if errorsstd.Is(err, sql.ErrNoRows) {
			return nil, banner_errors.ErrBannerNotFound
		}
		return nil, banner_errors.ErrTrashedBanner.WithInternal(err)
	}

	return &res, nil
}

func (r *bannerCommandRepository) Restore(ctx context.Context, bannerID int) (*db.Banner, error) {
	var res db.Banner
	err := r.db.GetContext(ctx, &res, restoreBanner, int32(bannerID))

	if err != nil {
		if errorsstd.Is(err, sql.ErrNoRows) {
			return nil, banner_errors.ErrBannerNotFound
		}
		return nil, banner_errors.ErrRestoreBanner.WithInternal(err)
	}

	return &res, nil
}

func (r *bannerCommandRepository) DeletePermanent(ctx context.Context, bannerID int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteBannerPermanently, int32(bannerID)); err != nil {
		return false, banner_errors.ErrDeleteBannerPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *bannerCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllBanners); err != nil {
		return false, banner_errors.ErrRestoreAllBanners.WithInternal(err)
	}
	return true, nil
}

func (r *bannerCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentBanners); err != nil {
		return false, banner_errors.ErrDeleteAllBanners.WithInternal(err)
	}
	return true, nil
}
