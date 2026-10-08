package seeder

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/database/schema"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const getShippingAddress = `SELECT
    shipping_address_id,
    order_id,
    alamat,
    provinsi,
    negara,
    kota,
    courier,
    shipping_method,
    shipping_cost,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM shipping_addresses
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR shipping_address_id::TEXT ILIKE '%' || $1 || '%'
        OR alamat ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const createShippingAddress = `INSERT INTO
    shipping_addresses (
        order_id,
        alamat,
        provinsi,
        negara,
        kota,
        courier,
        shipping_method,
        shipping_cost
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING
    shipping_address_id,
    order_id,
    alamat,
    provinsi,
    negara,
    kota,
    courier,
    shipping_method,
    shipping_cost,
    created_at,
    updated_at`

type shippingAddressSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewShippingAddressSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *shippingAddressSeeder {
	return &shippingAddressSeeder{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *shippingAddressSeeder) Seed() error {
	// Idempotency: skip when shipping addresses already exist.
	var existing []*db.GetShippingAddressRow
	err := r.db.SelectContext(r.ctx, &existing, getShippingAddress, "", int32(1), int32(0))
	if err == nil && len(existing) > 0 {
		r.logger.Debug("shipping addresses already seeded, skipping")
		return nil
	}

	addresses := []struct {
		OrderID        int32
		Alamat         string
		Provinsi       string
		Negara         string
		Kota           string
		Courier        string
		ShippingMethod string
		ShippingCost   float64
	}{
		{OrderID: 1, Alamat: "Jl. Sudirman No. 10", Provinsi: "DKI Jakarta", Negara: "Indonesia", Kota: "Jakarta", Courier: "JNE", ShippingMethod: "Reguler", ShippingCost: 12000},
		{OrderID: 2, Alamat: "Jl. Asia Afrika No. 20", Provinsi: "Jawa Barat", Negara: "Indonesia", Kota: "Bandung", Courier: "SiCepat", ShippingMethod: "Express", ShippingCost: 18000},
		{OrderID: 3, Alamat: "Jl. Diponegoro No. 15", Provinsi: "DI Yogyakarta", Negara: "Indonesia", Kota: "Yogyakarta", Courier: "J&T", ShippingMethod: "Reguler", ShippingCost: 15000},
		{OrderID: 4, Alamat: "Jl. Pemuda No. 9", Provinsi: "Jawa Tengah", Negara: "Indonesia", Kota: "Semarang", Courier: "TIKI", ShippingMethod: "Reguler", ShippingCost: 13000},
		{OrderID: 5, Alamat: "Jl. Basuki Rahmat No. 3", Provinsi: "Jawa Timur", Negara: "Indonesia", Kota: "Surabaya", Courier: "AnterAja", ShippingMethod: "Next Day", ShippingCost: 20000},
		{OrderID: 6, Alamat: "Jl. Sisingamangaraja No. 25", Provinsi: "Sumatera Utara", Negara: "Indonesia", Kota: "Medan", Courier: "JNE", ShippingMethod: "Reguler", ShippingCost: 16000},
		{OrderID: 7, Alamat: "Jl. Gatot Subroto No. 77", Provinsi: "Bali", Negara: "Indonesia", Kota: "Denpasar", Courier: "SiCepat", ShippingMethod: "Hemat", ShippingCost: 14000},
		{OrderID: 8, Alamat: "Jl. Gajah Mada No. 88", Provinsi: "Kalimantan Timur", Negara: "Indonesia", Kota: "Balikpapan", Courier: "TIKI", ShippingMethod: "Express", ShippingCost: 17000},
	}

	for _, address := range addresses {
		var row db.CreateShippingAddressRow
		if err := r.db.GetContext(r.ctx, &row, createShippingAddress,
			address.OrderID,
			address.Alamat,
			address.Provinsi,
			address.Negara,
			address.Kota,
			address.Courier,
			address.ShippingMethod,
			address.ShippingCost,
		); err != nil {
			r.logger.Error("failed to seed shipping address", zap.Error(err))
			return err
		}
	}

	r.logger.Info("shipping address successfully seeded")

	return nil
}
