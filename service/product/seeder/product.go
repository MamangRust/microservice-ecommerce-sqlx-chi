package seeder

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-product/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const getProductsSeed = `SELECT
    product_id,
    merchant_id,
    category_id,
    name,
    description,
    price,
    count_in_stock,
    brand,
    weight,
    rating,
    slug_product,
    image_product,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM products as p
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR p.name ILIKE '%' || $1 || '%'
        OR p.description ILIKE '%' || $1 || '%'
        OR p.brand ILIKE '%' || $1 || '%'
        OR p.slug_product ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const createProductSeed = `INSERT INTO
    products (
        merchant_id,
        category_id,
        name,
        description,
        price,
        count_in_stock,
        brand,
        weight,
        rating,
        slug_product,
        image_product
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11
    )
RETURNING
    product_id,
    merchant_id,
    category_id,
    name,
    description,
    price,
    count_in_stock,
    brand,
    weight,
    rating,
    slug_product,
    image_product,
    created_at,
    updated_at`

type productSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewProductSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *productSeeder {
	return &productSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *productSeeder) Seed() error {
	// Idempotency: skip when products already exist.
	var existing []*db.GetProductsRow
	err := r.db.SelectContext(r.ctx, &existing, getProductsSeed,
		"",
		int32(1),
		int32(0),
	)
	if err == nil && len(existing) > 0 {
		r.logger.Debug("products already seeded, skipping")
		return nil
	}

	products := []db.CreateProductParams{
		{
			MerchantID:   1,
			CategoryID:   1,
			Name:         "Smartphone Galaxy X",
			Description:  toStringPtr("Smartphone dengan performa tinggi dan kamera canggih."),
			Price:        4500000,
			CountInStock: 20,
			Brand:        toStringPtr("Samsung"),
			Weight:       toInt32Ptr(300),
			Rating:       toFloat64Ptr(4.5),
			SlugProduct:  toStringPtr("smartphone-galaxy-x"),
			ImageProduct: toStringPtr("galaxy-x.jpg"),
		},
		{
			MerchantID:   2,
			CategoryID:   2,
			Name:         "Facial Cleanser Glow",
			Description:  toStringPtr("Pembersih wajah dengan formula ringan untuk semua jenis kulit."),
			Price:        75000,
			CountInStock: 100,
			Brand:        toStringPtr("GlowCare"),
			Weight:       toInt32Ptr(150),
			Rating:       toFloat64Ptr(4.2),
			SlugProduct:  toStringPtr("facial-cleanser-glow"),
			ImageProduct: toStringPtr("cleanser.jpg"),
		},
		{
			MerchantID:   3,
			CategoryID:   3,
			Name:         "Blender Serbaguna",
			Description:  toStringPtr("Blender 3-in-1 untuk keperluan dapur sehari-hari."),
			Price:        350000,
			CountInStock: 50,
			Brand:        toStringPtr("Maspion"),
			Weight:       toInt32Ptr(2000),
			Rating:       toFloat64Ptr(4.0),
			SlugProduct:  toStringPtr("blender-serbaguna"),
			ImageProduct: toStringPtr("blender.jpg"),
		},
		{
			MerchantID:   4,
			CategoryID:   4,
			Name:         "Paket Popok Bayi Premium",
			Description:  toStringPtr("Popok bayi dengan teknologi anti bocor dan lembut di kulit."),
			Price:        120000,
			CountInStock: 70,
			Brand:        toStringPtr("BabySoft"),
			Weight:       toInt32Ptr(1000),
			Rating:       toFloat64Ptr(4.6),
			SlugProduct:  toStringPtr("popok-premium"),
			ImageProduct: toStringPtr("popok.jpg"),
		},
		{
			MerchantID:   5,
			CategoryID:   5,
			Name:         "Matras Yoga Premium",
			Description:  toStringPtr("Matras anti slip dengan ketebalan ideal untuk yoga dan fitness."),
			Price:        220000,
			CountInStock: 40,
			Brand:        toStringPtr("FitZone"),
			Weight:       toInt32Ptr(700),
			Rating:       toFloat64Ptr(4.4),
			SlugProduct:  toStringPtr("matras-yoga-premium"),
			ImageProduct: toStringPtr("matras.jpg"),
		},
		{
			MerchantID:   6,
			CategoryID:   6,
			Name:         "Snack Kentang Balado",
			Description:  toStringPtr("Cemilan kentang renyah dengan rasa balado khas."),
			Price:        18000,
			CountInStock: 200,
			Brand:        toStringPtr("Snacky"),
			Weight:       toInt32Ptr(100),
			Rating:       toFloat64Ptr(4.1),
			SlugProduct:  toStringPtr("kentang-balado"),
			ImageProduct: toStringPtr("snack.jpg"),
		},
		{
			MerchantID:   7,
			CategoryID:   7,
			Name:         "Controller PS5 DualSense",
			Description:  toStringPtr("Stik PS5 dengan fitur haptic feedback dan adaptive triggers."),
			Price:        999000,
			CountInStock: 30,
			Brand:        toStringPtr("Sony"),
			Weight:       toInt32Ptr(450),
			Rating:       toFloat64Ptr(4.8),
			SlugProduct:  toStringPtr("controller-ps5"),
			ImageProduct: toStringPtr("ps5-controller.jpg"),
		},
		{
			MerchantID:   8,
			CategoryID:   8,
			Name:         "Oli Motor Full Synthetic",
			Description:  toStringPtr("Oli mesin motor dengan perlindungan maksimal dan efisiensi tinggi."),
			Price:        95000,
			CountInStock: 60,
			Brand:        toStringPtr("Motul"),
			Weight:       toInt32Ptr(1000),
			Rating:       toFloat64Ptr(4.3),
			SlugProduct:  toStringPtr("oli-motor-synthetic"),
			ImageProduct: toStringPtr("oli.jpg"),
		},
	}

	for _, product := range products {
		var created db.CreateProductRow
		if err := r.db.GetContext(r.ctx, &created, createProductSeed,
			product.MerchantID,
			product.CategoryID,
			product.Name,
			product.Description,
			product.Price,
			product.CountInStock,
			product.Brand,
			product.Weight,
			product.Rating,
			product.SlugProduct,
			product.ImageProduct,
		); err != nil {
			r.logger.Error("failed to seed product", zap.Error(err))
			return err
		}
	}

	r.logger.Info("product successfully seeded")

	return nil
}
