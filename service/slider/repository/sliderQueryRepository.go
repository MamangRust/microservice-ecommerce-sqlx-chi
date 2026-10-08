package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-slider/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/slider_errors"
	"github.com/jmoiron/sqlx"
)

const getSliders = `SELECT
    slider_id,
    name,
    image,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM sliders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getSlidersActive = `SELECT
    slider_id,
    name,
    image,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM sliders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getSlidersTrashed = `SELECT
    slider_id,
    name,
    image,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM sliders
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getSliderByID = `SELECT
    slider_id,
    name,
    image,
    created_at,
    updated_at
FROM sliders
WHERE
    slider_id = $1
    AND deleted_at IS NULL`

type sliderQueryRepository struct {
	db *sqlx.DB
}

func NewSliderQueryRepository(db *sqlx.DB) *sliderQueryRepository {
	return &sliderQueryRepository{
		db: db,
	}
}

func (r *sliderQueryRepository) FindAll(ctx context.Context, req *requests.FindAllSlider) ([]*db.GetSlidersRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetSlidersRow
	err := r.db.SelectContext(ctx, &rows, getSliders, req.Search, int32(req.PageSize), int32(offset))
	if err != nil {
		return nil, slider_errors.ErrFindAllSliders
	}

	return rows, nil
}

func (r *sliderQueryRepository) FindActive(ctx context.Context, req *requests.FindAllSlider) ([]*db.GetSlidersActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetSlidersActiveRow
	err := r.db.SelectContext(ctx, &rows, getSlidersActive, req.Search, int32(req.PageSize), int32(offset))
	if err != nil {
		return nil, slider_errors.ErrFindActiveSliders
	}

	return rows, nil
}

func (r *sliderQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllSlider) ([]*db.GetSlidersTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetSlidersTrashedRow
	err := r.db.SelectContext(ctx, &rows, getSlidersTrashed, req.Search, int32(req.PageSize), int32(offset))
	if err != nil {
		return nil, slider_errors.ErrFindTrashedSliders
	}

	return rows, nil
}

func (r *sliderQueryRepository) FindByID(ctx context.Context, slider_id int) (*db.GetSliderByIDRow, error) {
	var row db.GetSliderByIDRow
	err := r.db.GetContext(ctx, &row, getSliderByID, int32(slider_id))
	if err != nil {
		return nil, slider_errors.ErrFindSliderByID
	}

	return &row, nil
}
