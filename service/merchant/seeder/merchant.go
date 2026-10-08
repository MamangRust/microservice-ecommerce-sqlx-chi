package seeder

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

type merchantSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewMerchantSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *merchantSeeder {
	return &merchantSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

// merchantSeed mirrors the columns of the createMerchant query.
type merchantSeed struct {
	userID       int32
	name         string
	description  *string
	address      *string
	contactEmail *string
	contactPhone *string
	status       string
}

// getMerchants is copied verbatim from database/query/merchants.sql (GetMerchants)
// and used for the idempotency check before seeding.
const getMerchants = `SELECT
    merchant_id,
    user_id,
    name,
    description,
    address,
    contact_email,
    contact_phone,
    status,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM merchants
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR name ILIKE '%' || $1 || '%'
        OR contact_email ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET $3`

// createMerchant is copied verbatim from database/query/merchants.sql (CreateMerchant).
const createMerchant = `INSERT INTO
    merchants (
        user_id,
        name,
        description,
        address,
        contact_email,
        contact_phone,
        status
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)`

func (r *merchantSeeder) Seed() error {
	// Idempotency: skip when merchants already exist.
	var existing []struct {
		MerchantID int32 `db:"merchant_id"`
	}
	err := r.db.SelectContext(r.ctx, &existing, getMerchants, "", int32(1), int32(0))
	if err == nil && len(existing) > 0 {
		r.logger.Debug("merchants already seeded, skipping")
		return nil
	}

	merchants := []merchantSeed{
		{
			userID:       1,
			name:         "Elektronik Store",
			description:  toStringPtr("Toko elektronik terpercaya dengan berbagai produk gadget dan aksesoris."),
			address:      toStringPtr("Jl. Teknologi No.1, Jakarta"),
			contactEmail: toStringPtr("support@elektronikstore.com"),
			contactPhone: toStringPtr("081234567890"),
			status:       "active",
		},
		{
			userID:       2,
			name:         "Kecantikan Sehat",
			description:  toStringPtr("Produk skincare dan kesehatan pilihan."),
			address:      toStringPtr("Jl. Kesehatan No.5, Bandung"),
			contactEmail: toStringPtr("cs@kecantikansehat.com"),
			contactPhone: toStringPtr("082345678901"),
			status:       "active",
		},
		{
			userID:       3,
			name:         "Rumah Indah",
			description:  toStringPtr("Peralatan rumah tangga berkualitas dan estetik."),
			address:      toStringPtr("Jl. Rumah No.12, Surabaya"),
			contactEmail: toStringPtr("info@rumahindah.com"),
			contactPhone: toStringPtr("083456789012"),
			status:       "active",
		},
		{
			userID:       4,
			name:         "Mom & Baby Care",
			description:  toStringPtr("Semua kebutuhan ibu dan bayi ada di sini."),
			address:      toStringPtr("Jl. Keluarga No.7, Depok"),
			contactEmail: toStringPtr("support@momandbaby.com"),
			contactPhone: toStringPtr("084567890123"),
			status:       "active",
		},
		{
			userID:       5,
			name:         "Sport Zone",
			description:  toStringPtr("Perlengkapan olahraga dan outdoor terlengkap."),
			address:      toStringPtr("Jl. Atletik No.3, Yogyakarta"),
			contactEmail: toStringPtr("halo@sportzone.com"),
			contactPhone: toStringPtr("085678901234"),
			status:       "active",
		},
		{
			userID:       6,
			name:         "Fresh Mart",
			description:  toStringPtr("Toko makanan dan minuman segar dan kemasan."),
			address:      toStringPtr("Jl. Pasar No.10, Semarang"),
			contactEmail: toStringPtr("fresh@mart.com"),
			contactPhone: toStringPtr("086789012345"),
			status:       "active",
		},
		{
			userID:       7,
			name:         "Gamer Heaven",
			description:  toStringPtr("Game, console, dan aksesori lengkap untuk gamers."),
			address:      toStringPtr("Jl. Game No.8, Bekasi"),
			contactEmail: toStringPtr("gamer@heaven.com"),
			contactPhone: toStringPtr("087890123456"),
			status:       "active",
		},
		{
			userID:       8,
			name:         "AutoParts Store",
			description:  toStringPtr("Toko perlengkapan otomotif terpercaya."),
			address:      toStringPtr("Jl. Otomotif No.6, Medan"),
			contactEmail: toStringPtr("service@autoparts.com"),
			contactPhone: toStringPtr("088901234567"),
			status:       "active",
		},
	}

	for _, merchant := range merchants {
		if _, err := r.db.ExecContext(r.ctx, createMerchant,
			merchant.userID,
			merchant.name,
			merchant.description,
			merchant.address,
			merchant.contactEmail,
			merchant.contactPhone,
			merchant.status,
		); err != nil {
			r.logger.Error("failed to seed merchant", zap.Error(err))
			return err
		}
	}

	r.logger.Info("merchant succesfully seeded")

	return nil
}
