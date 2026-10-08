package seeder

import (
	"context"

	reviewdetaildb "github.com/MamangRust/microservice-ecommerce-grpc-review-detail/database/schema"
	db "github.com/MamangRust/microservice-ecommerce-grpc-review/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const createReview = `INSERT INTO
    reviews (
        user_id,
        product_id,
        name,
        comment,
        rating
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at`

// getReviewsAny matches the reviews search query with an empty search term
// (matches everything), used for idempotency checks.
const getReviewsAny = `SELECT
    review_id,
    user_id,
    product_id,
    name,
    comment,
    rating,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM reviews
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR review_id::TEXT ILIKE '%' || $1 || '%'
        OR name ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

// createReviewDetail inserts one review_details row into the review_detail
// service database (inlined from review_detail's database/query/reviews_detail.sql).
const createReviewDetail = `INSERT INTO
    review_details (review_id, type, url, caption)
VALUES ($1, $2, $3, $4)
RETURNING
    review_detail_id,
    review_id,
    type,
    url,
    caption,
    created_at,
    updated_at`

// reviewSeeder seeds reviews (review service DB) and their review details
// (review_detail service DB), so it needs both connections.
type reviewSeeder struct {
	reviewDB       *sqlx.DB
	reviewDetailDB *sqlx.DB
	ctx            context.Context
	logger         logger.LoggerInterface
}

func NewReviewSeeder(reviewDB *sqlx.DB, reviewDetailDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *reviewSeeder {
	return &reviewSeeder{
		reviewDB:       reviewDB,
		reviewDetailDB: reviewDetailDB,
		ctx:            ctx,
		logger:         logger,
	}
}

type seedReview struct {
	UserID    int32
	ProductID int32
	Name      string
	Comment   string
	Rating    int32
}

func (r *reviewSeeder) Seed() error {
	// Idempotency: skip when reviews already exist.
	var existing []*db.GetReviewsRow
	if err := r.reviewDB.SelectContext(r.ctx, &existing, getReviewsAny, "", int32(1), int32(0)); err == nil && len(existing) > 0 {
		r.logger.Debug("reviews already seeded, skipping")
		return nil
	}

	reviews := []seedReview{
		{UserID: 1, ProductID: 1, Name: "John", Comment: "Produk bagus!", Rating: 5},
		{UserID: 2, ProductID: 2, Name: "Anna", Comment: "Sangat puas dengan kualitasnya.", Rating: 4},
		{UserID: 3, ProductID: 3, Name: "Budi", Comment: "Cukup oke untuk harga segini.", Rating: 3},
		{UserID: 4, ProductID: 4, Name: "Siti", Comment: "Pengiriman cepat dan aman.", Rating: 4},
		{UserID: 5, ProductID: 5, Name: "Rina", Comment: "Tidak sesuai ekspektasi.", Rating: 2},
		{UserID: 6, ProductID: 6, Name: "Agus", Comment: "Top, pasti beli lagi!", Rating: 5},
		{UserID: 7, ProductID: 7, Name: "Dian", Comment: "Cocok untuk hadiah.", Rating: 4},
		{UserID: 8, ProductID: 8, Name: "Made", Comment: "Kualitas standar saja.", Rating: 3},
	}

	for _, review := range reviews {
		var createdReview db.CreateReviewRow
		err := r.reviewDB.GetContext(r.ctx, &createdReview, createReview,
			review.UserID,
			review.ProductID,
			review.Name,
			review.Comment,
			review.Rating,
		)
		if err != nil {
			r.logger.Error("failed to create review", zap.Error(err))
			return err
		}

		var createdDetail reviewdetaildb.ReviewDetail
		err = r.reviewDetailDB.GetContext(r.ctx, &createdDetail, createReviewDetail,
			createdReview.ReviewID,
			"photo",
			"https://example.com/review_"+review.Name+".jpg",
			toStringPtr("Foto review oleh "+review.Name),
		)
		if err != nil {
			r.logger.Error("failed to create review detail", zap.Error(err))
			return err
		}
	}

	r.logger.Info("review & review detail successfully seeded")

	return nil
}
