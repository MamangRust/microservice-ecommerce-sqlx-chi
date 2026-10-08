package seeder

import (
	"context"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-grpc-banner/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
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

type bannerSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewBannerSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *bannerSeeder {
	return &bannerSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *bannerSeeder) Seed() error {
	// Idempotency: skip when banners already exist.
	var existing []db.GetBannersRow
	err := r.db.SelectContext(r.ctx, &existing, getBanners, "", int32(1), int32(0))
	if err == nil && len(existing) > 0 {
		r.logger.Debug("banners already seeded, skipping")
		return nil
	}

	banners := []db.CreateBannerParams{
		{
			Name:      "Banner 1",
			StartDate: toDate(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("08:00")),
			EndTime:   toTime(parseTime("16:00")),
			IsActive:  toBoolPtr(true),
		},
		{
			Name:      "Banner 2",
			StartDate: toDate(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("09:00")),
			EndTime:   toTime(parseTime("17:00")),
			IsActive:  toBoolPtr(true),
		},
		{
			Name:      "Banner 3",
			StartDate: toDate(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 3, 31, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("10:00")),
			EndTime:   toTime(parseTime("18:00")),
			IsActive:  toBoolPtr(false),
		},
		{
			Name:      "Banner 4",
			StartDate: toDate(time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 4, 30, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("07:00")),
			EndTime:   toTime(parseTime("15:00")),
			IsActive:  toBoolPtr(true),
		},
		{
			Name:      "Banner 5",
			StartDate: toDate(time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 5, 31, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("06:00")),
			EndTime:   toTime(parseTime("14:00")),
			IsActive:  toBoolPtr(false),
		},
		{
			Name:      "Banner 6",
			StartDate: toDate(time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("12:00")),
			EndTime:   toTime(parseTime("20:00")),
			IsActive:  toBoolPtr(true),
		},
		{
			Name:      "Banner 7",
			StartDate: toDate(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 7, 31, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("08:30")),
			EndTime:   toTime(parseTime("16:30")),
			IsActive:  toBoolPtr(true),
		},
		{
			Name:      "Banner 8",
			StartDate: toDate(time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)),
			EndDate:   toDate(time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC)),
			StartTime: toTime(parseTime("09:30")),
			EndTime:   toTime(parseTime("17:30")),
			IsActive:  toBoolPtr(false),
		},
	}

	for _, banner := range banners {
		var created db.CreateBannerRow
		if err := r.db.GetContext(r.ctx, &created, createBanner,
			banner.Name,
			banner.StartDate,
			banner.EndDate,
			banner.StartTime,
			banner.EndTime,
			banner.IsActive,
		); err != nil {
			r.logger.Error("Failed to insert banner", zap.Error(err))
			return err
		}
	}

	r.logger.Info("banner successfully seeded")
	return nil
}

func parseTime(t string) time.Time {
	parsed, _ := time.Parse("15:04", t)
	return parsed
}
