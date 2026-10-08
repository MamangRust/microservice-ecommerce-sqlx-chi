package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-shipping-address/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shippingaddress_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/shipping_address_errors"
	"github.com/jmoiron/sqlx"
)

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

const updateShippingAddress = `UPDATE shipping_addresses
SET
    alamat = $2,
    provinsi = $3,
    negara = $4,
    kota = $5,
    courier = $6,
    shipping_method = $7,
    shipping_cost = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE
    shipping_address_id = $1
    AND deleted_at IS NULL
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

const trashShippingAddress = `UPDATE shipping_addresses
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    shipping_address_id = $1
    AND deleted_at IS NULL
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
    updated_at,
    deleted_at`

const restoreShippingAddress = `UPDATE shipping_addresses
SET
    deleted_at = NULL
WHERE
    shipping_address_id = $1
    AND deleted_at IS NOT NULL
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
    updated_at,
    deleted_at`

const deleteShippingAddressPermanently = `DELETE FROM shipping_addresses
WHERE
    shipping_address_id = $1
    AND deleted_at IS NOT NULL`

const deleteShippingAddressByOrderPermanent = `DELETE FROM shipping_addresses
WHERE
    order_id = $1`

const restoreAllShippingAddress = `UPDATE shipping_addresses
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentShippingAddress = `-- Permanently removes all trashed addresses. The orders cross-DB subquery was
-- removed: orders live in the order service DB (1 DB per service), so cleanup
-- of addresses whose parent order is trashed happens per-order via
-- DeleteShippingAddressByOrderPermanent instead.
DELETE FROM shipping_addresses
WHERE
    deleted_at IS NOT NULL`

type shippingAddressCommandRepository struct {
	db *sqlx.DB
}

func NewShippingAddressCommandRepository(db *sqlx.DB) *shippingAddressCommandRepository {
	return &shippingAddressCommandRepository{
		db: db,
	}
}

func (r *shippingAddressCommandRepository) Create(ctx context.Context, request *requests.CreateShippingAddressRequest) (*db.CreateShippingAddressRow, error) {
	var orderID int32
	if request.OrderID != nil {
		orderID = int32(*request.OrderID)
	}

	var address db.CreateShippingAddressRow
	err := r.db.GetContext(ctx, &address, createShippingAddress,
		orderID,
		request.Alamat,
		request.Provinsi,
		request.Negara,
		request.Kota,
		request.Courier,
		request.ShippingMethod,
		float64(request.ShippingCost),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return nil, shippingaddress_errors.ErrCreateShippingAddress
	}

	return &address, nil
}

func (r *shippingAddressCommandRepository) Update(ctx context.Context, request *requests.UpdateShippingAddressRequest) (*db.UpdateShippingAddressRow, error) {
	var res db.UpdateShippingAddressRow
	err := r.db.GetContext(ctx, &res, updateShippingAddress,
		int32(*request.ShippingID),
		request.Alamat,
		request.Provinsi,
		request.Negara,
		request.Kota,
		request.Courier,
		request.ShippingMethod,
		float64(request.ShippingCost),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return nil, shippingaddress_errors.ErrUpdateShippingAddress
	}

	return &res, nil
}

func (r *shippingAddressCommandRepository) Trash(ctx context.Context, shipping_id int) (*db.ShippingAddress, error) {
	var res db.ShippingAddress
	err := r.db.GetContext(ctx, &res, trashShippingAddress, int32(shipping_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return nil, shippingaddress_errors.ErrTrashShippingAddress
	}

	return &res, nil
}

func (r *shippingAddressCommandRepository) Restore(ctx context.Context, shipping_id int) (*db.ShippingAddress, error) {
	var res db.ShippingAddress
	err := r.db.GetContext(ctx, &res, restoreShippingAddress, int32(shipping_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return nil, shippingaddress_errors.ErrRestoreShippingAddress
	}

	return &res, nil
}

func (r *shippingAddressCommandRepository) DeletePermanent(ctx context.Context, shipping_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteShippingAddressPermanently, int32(shipping_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return false, shippingaddress_errors.ErrDeleteShippingAddressPermanent
	}

	return true, nil
}

func (r *shippingAddressCommandRepository) DeleteByOrderIDPermanent(ctx context.Context, order_id int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteShippingAddressByOrderPermanent, int32(order_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return false, shippingaddress_errors.ErrDeleteShippingAddressPermanent
	}

	return true, nil
}

func (r *shippingAddressCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllShippingAddress)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return false, shippingaddress_errors.ErrRestoreAllShippingAddresses
	}
	return true, nil
}

func (r *shippingAddressCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentShippingAddress)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, shippingaddress_errors.ErrShippingAddressNotFound
		}
		return false, shippingaddress_errors.ErrDeleteAllPermanentShippingAddress
	}
	return true, nil
}
