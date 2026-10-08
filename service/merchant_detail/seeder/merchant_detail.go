package seeder

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const getMerchantDetailsSeed = `SELECT
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

const createMerchantDetailSeed = `INSERT INTO
    merchant_details (
        merchant_id,
        display_name,
        cover_image_url,
        logo_url,
        short_description,
        website_url
    )
VALUES ($1, $2, $3, $4, $5, $6)`

const createMerchantSocialMediaLinkSeed = `INSERT INTO
    merchant_social_media_links (
        merchant_detail_id,
        platform,
        url
    )
VALUES ($1, $2, $3)`

type merchantDetailSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewMerchantDetailSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *merchantDetailSeeder {
	return &merchantDetailSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *merchantDetailSeeder) Seed() error {
	// Idempotency: skip when merchant details already exist.
	existing := []struct{}{}
	if err := r.db.SelectContext(r.ctx, &existing, getMerchantDetailsSeed, "", int32(1), int32(0)); err == nil && len(existing) > 0 {
		r.logger.Debug("merchant details already seeded, skipping")
		return nil
	}

	details := []dbSeedParams{
		{
			MerchantID:       1,
			DisplayName:      toStringPtr("Techno Store"),
			CoverImageUrl:    toStringPtr("cover/techno.jpg"),
			LogoUrl:          toStringPtr("logo/techno.png"),
			ShortDescription: toStringPtr("Pusat elektronik terpercaya sejak 2010"),
			WebsiteUrl:       toStringPtr("https://technostore.com"),
		},
		{
			MerchantID:       2,
			DisplayName:      toStringPtr("Glow Beauty"),
			CoverImageUrl:    toStringPtr("cover/beauty.jpg"),
			LogoUrl:          toStringPtr("logo/beauty.png"),
			ShortDescription: toStringPtr("Produk kecantikan alami dan aman"),
			WebsiteUrl:       toStringPtr("https://glowbeauty.id"),
		},
		{
			MerchantID:       3,
			DisplayName:      toStringPtr("Dapur Sehat"),
			CoverImageUrl:    toStringPtr("cover/dapur.jpg"),
			LogoUrl:          toStringPtr("logo/dapur.png"),
			ShortDescription: toStringPtr("Makanan sehat dan organik"),
			WebsiteUrl:       toStringPtr("https://dapsehat.id"),
		},
		{
			MerchantID:       4,
			DisplayName:      toStringPtr("Gadget Hub"),
			CoverImageUrl:    toStringPtr("cover/gadget.jpg"),
			LogoUrl:          toStringPtr("logo/gadget.png"),
			ShortDescription: toStringPtr("Semua tentang gadget terbaru"),
			WebsiteUrl:       toStringPtr("https://gadgethub.com"),
		},
		{
			MerchantID:       5,
			DisplayName:      toStringPtr("Bayi Ceria"),
			CoverImageUrl:    toStringPtr("cover/bayi.jpg"),
			LogoUrl:          toStringPtr("logo/bayi.png"),
			ShortDescription: toStringPtr("Produk terbaik untuk si kecil"),
			WebsiteUrl:       toStringPtr("https://bayiceria.id"),
		},
		{
			MerchantID:       6,
			DisplayName:      toStringPtr("Toko Sehat"),
			CoverImageUrl:    toStringPtr("cover/sehat.jpg"),
			LogoUrl:          toStringPtr("logo/sehat.png"),
			ShortDescription: toStringPtr("Peralatan olahraga lengkap"),
			WebsiteUrl:       toStringPtr("https://tokosehat.id"),
		},
		{
			MerchantID:       7,
			DisplayName:      toStringPtr("Game World"),
			CoverImageUrl:    toStringPtr("cover/game.jpg"),
			LogoUrl:          toStringPtr("logo/game.png"),
			ShortDescription: toStringPtr("Konsol dan game terbaik"),
			WebsiteUrl:       toStringPtr("https://gameworld.com"),
		},
		{
			MerchantID:       8,
			DisplayName:      toStringPtr("Otomotif Mart"),
			CoverImageUrl:    toStringPtr("cover/otomotif.jpg"),
			LogoUrl:          toStringPtr("logo/otomotif.png"),
			ShortDescription: toStringPtr("Aksesori kendaraan terpercaya"),
			WebsiteUrl:       toStringPtr("https://otomotifmart.com"),
		},
	}

	for i, detail := range details {
		if _, err := r.db.ExecContext(r.ctx, createMerchantDetailSeed,
			detail.MerchantID,
			detail.DisplayName,
			detail.CoverImageUrl,
			detail.LogoUrl,
			detail.ShortDescription,
			detail.WebsiteUrl,
		); err != nil {
			r.logger.Error("failed to seed merchant detail", zap.Error(err))
			return err
		}

		merchantDetailID := int32(i + 1)
		socialMedia := []dbSeedSocialParams{
			{
				MerchantDetailID: merchantDetailID,
				Platform:         "Facebook",
				Url:              "https://www.facebook.com/merchant" + fmt.Sprint(merchantDetailID),
			},
			{
				MerchantDetailID: merchantDetailID,
				Platform:         "Instagram",
				Url:              "https://www.instagram.com/merchant" + fmt.Sprint(merchantDetailID),
			},
			{
				MerchantDetailID: merchantDetailID,
				Platform:         "Twitter",
				Url:              "https://www.twitter.com/merchant" + fmt.Sprint(merchantDetailID),
			},
		}

		for _, sm := range socialMedia {
			if _, err := r.db.ExecContext(r.ctx, createMerchantSocialMediaLinkSeed,
				sm.MerchantDetailID,
				sm.Platform,
				sm.Url,
			); err != nil {
				r.logger.Error("failed to seed merchant social media link", zap.Error(err))
				return err
			}
		}
	}

	r.logger.Info("merchant detail & merchant social link successfully seeded")

	return nil
}
