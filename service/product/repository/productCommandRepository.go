package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-product/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/product_errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const createProduct = `INSERT INTO
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

const updateProduct = `UPDATE products
SET
    category_id = $2,
    name = $3,
    description = $4,
    price = $5,
    count_in_stock = $6,
    brand = $7,
    weight = $8,
    rating = $9,
    slug_product = $10,
    image_product = $11,
    updated_at = CURRENT_TIMESTAMP
WHERE
    product_id = $1
    AND deleted_at IS NULL
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

const updateProductCountStock = `UPDATE products
SET
    count_in_stock = $2
WHERE
    product_id = $1
    AND deleted_at IS NULL
    AND $2 >= 0
RETURNING
    product_id,
    count_in_stock`

const adjustProductStock = `WITH existing_adjustment AS (
    SELECT product_id, delta
    FROM product_stock_adjustments psa
    WHERE psa.operation_id = $1
      AND psa.product_id = $2
      AND psa.delta = $3
), new_adjustment AS (
    INSERT INTO product_stock_adjustments (operation_id, product_id, delta)
    SELECT $1, $2, $3
    WHERE NOT EXISTS (SELECT 1 FROM existing_adjustment)
      AND EXISTS (
          SELECT 1
          FROM products
          WHERE product_id = $2
            AND deleted_at IS NULL
            AND count_in_stock + $3 >= 0
      )
    ON CONFLICT (operation_id) DO NOTHING
    RETURNING product_id, delta
), applied AS (
    UPDATE products p
    SET count_in_stock = p.count_in_stock + a.delta,
        updated_at = CURRENT_TIMESTAMP
    FROM new_adjustment a
    WHERE p.product_id = a.product_id
      AND p.deleted_at IS NULL
    RETURNING p.product_id, p.count_in_stock
)
SELECT a.product_id, a.count_in_stock
FROM applied a
UNION ALL
SELECT p.product_id, p.count_in_stock
FROM products p
JOIN existing_adjustment e ON e.product_id = p.product_id
WHERE p.deleted_at IS NULL`

const trashProduct = `UPDATE products
SET
    deleted_at = current_timestamp
WHERE
    product_id = $1
    AND deleted_at IS NULL
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

const restoreProduct = `UPDATE products
SET
    deleted_at = NULL
WHERE
    product_id = $1
    AND deleted_at IS NOT NULL
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

const deleteProductPermanently = `DELETE FROM products
WHERE
    product_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllProducts = `UPDATE products
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentProducts = `DELETE FROM products WHERE deleted_at IS NOT NULL`

type productCommandRepository struct {
	db *sqlx.DB
}

func NewProductCommandRepository(db *sqlx.DB) *productCommandRepository {
	return &productCommandRepository{
		db: db,
	}
}

func (r *productCommandRepository) Create(ctx context.Context, request *requests.CreateProductRequest) (*db.CreateProductRow, error) {
	var product db.CreateProductRow
	err := r.db.GetContext(ctx, &product, createProduct,
		int32(request.MerchantID),
		int32(request.CategoryID),
		request.Name,
		stringPtr(request.Description),
		int32(request.Price),
		int32(request.CountInStock),
		stringPtr(request.Brand),
		int32Ptr(request.Weight),
		nil,
		request.SlugProduct,
		stringPtr(request.ImageProduct),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrCreateProduct.WithInternal(err)
	}

	return &product, nil
}

func (r *productCommandRepository) Update(ctx context.Context, request *requests.UpdateProductRequest) (*db.UpdateProductRow, error) {
	var product db.UpdateProductRow
	err := r.db.GetContext(ctx, &product, updateProduct,
		int32(*request.ProductID),
		int32(request.CategoryID),
		request.Name,
		stringPtr(request.Description),
		int32(request.Price),
		int32(request.CountInStock),
		stringPtr(request.Brand),
		int32Ptr(request.Weight),
		nil,
		request.SlugProduct,
		stringPtr(request.ImageProduct),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrUpdateProduct.WithInternal(err)
	}

	return &product, nil
}

func (r *productCommandRepository) UpdateProductCountStock(ctx context.Context, product_id int, stock int) (*db.UpdateProductCountStockRow, error) {
	var row db.UpdateProductCountStockRow
	err := r.db.GetContext(ctx, &row, updateProductCountStock,
		int32(product_id),
		int32(stock),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrProductInternal.WithInternal(err)
	}

	return &row, nil
}

func (r *productCommandRepository) AdjustProductStock(ctx context.Context, product_id int, delta int, operationID string) (*db.AdjustProductStockRow, error) {
	var row db.AdjustProductStockRow
	err := r.db.GetContext(ctx, &row, adjustProductStock,
		operationID,
		int32(product_id),
		int32(delta),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrUpdateProductCountStock.WithInternal(err)
	}

	return &row, nil
}

func (r *productCommandRepository) Trash(ctx context.Context, product_id int) (*db.TrashProductRow, error) {
	var row db.TrashProductRow
	err := r.db.GetContext(ctx, &row, trashProduct, int32(product_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrTrashedProduct.WithInternal(err)
	}

	return &row, nil
}

func (r *productCommandRepository) Restore(ctx context.Context, product_id int) (*db.RestoreProductRow, error) {
	var row db.RestoreProductRow
	err := r.db.GetContext(ctx, &row, restoreProduct, int32(product_id))

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound
		}
		return nil, product_errors.ErrRestoreProduct.WithInternal(err)
	}

	return &row, nil
}

func (r *productCommandRepository) DeletePermanent(ctx context.Context, product_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteProductPermanently, int32(product_id)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete product while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, product_errors.ErrProductNotFound
		}
		return false, product_errors.ErrDeleteProductPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *productCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllProducts); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, product_errors.ErrProductNotFound
		}
		return false, product_errors.ErrRestoreAllProducts.WithInternal(err)
	}

	return true, nil
}

func (r *productCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentProducts); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, shared_errors.NewConflictError("cannot permanently delete products while related records exist").WithInternal(err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return false, product_errors.ErrProductNotFound
		}
		return false, product_errors.ErrDeleteAllProducts.WithInternal(err)
	}

	return true, nil
}
