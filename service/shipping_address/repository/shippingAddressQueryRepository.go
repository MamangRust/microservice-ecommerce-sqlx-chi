package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shippingaddress_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/shipping_address_errors"
	"github.com/jmoiron/sqlx"
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

const getShippingAddressActive = `SELECT
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
    deleted_at,
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

const getShippingAddressTrashed = `SELECT
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
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM shipping_addresses
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR shipping_address_id::TEXT ILIKE '%' || $1 || '%'
        OR alamat ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const getShippingByID = `SELECT
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
    updated_at
FROM shipping_addresses
WHERE
    shipping_address_id = $1
    AND deleted_at IS NULL`

const getShippingAddressByOrderID = `SELECT
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
    updated_at
FROM shipping_addresses
WHERE
    order_id = $1
    AND deleted_at IS NULL`

type shippingAddressQueryRepository struct {
	db *sqlx.DB
}

func NewShippingAddressQueryRepository(db *sqlx.DB) *shippingAddressQueryRepository {
	return &shippingAddressQueryRepository{
		db: db,
	}
}

func (r *shippingAddressQueryRepository) FindAll(ctx context.Context, req *requests.FindAllShippingAddress) ([]*db.GetShippingAddressRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetShippingAddressRow
	err := r.db.SelectContext(ctx, &res, getShippingAddress,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, shippingaddress_errors.ErrFindAllShippingAddress
	}

	return res, nil
}

func (r *shippingAddressQueryRepository) FindActive(ctx context.Context, req *requests.FindAllShippingAddress) ([]*db.GetShippingAddressActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetShippingAddressActiveRow
	err := r.db.SelectContext(ctx, &res, getShippingAddressActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, shippingaddress_errors.ErrFindActiveShippingAddress
	}

	return res, nil
}

func (r *shippingAddressQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllShippingAddress) ([]*db.GetShippingAddressTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetShippingAddressTrashedRow
	err := r.db.SelectContext(ctx, &res, getShippingAddressTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)

	if err != nil {
		return nil, shippingaddress_errors.ErrFindTrashedShippingAddress
	}

	return res, nil
}

func (r *shippingAddressQueryRepository) FindByID(ctx context.Context, shipping_id int) (*db.GetShippingByIDRow, error) {
	var res db.GetShippingByIDRow
	err := r.db.GetContext(ctx, &res, getShippingByID, int32(shipping_id))

	if err != nil {
		return nil, shippingaddress_errors.ErrFindShippingAddressByID
	}

	return &res, nil
}

func (r *shippingAddressQueryRepository) FindByOrder(ctx context.Context, order_id int) (*db.GetShippingAddressByOrderIDRow, error) {
	var res db.GetShippingAddressByOrderIDRow
	err := r.db.GetContext(ctx, &res, getShippingAddressByOrderID, int32(order_id))

	if err != nil {
		return nil, shippingaddress_errors.ErrFindShippingAddressByOrder
	}

	return &res, nil
}
