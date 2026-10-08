package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	categoryadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/category"
	db "github.com/MamangRust/microservice-ecommerce-grpc-product/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/product_errors"
	"github.com/jmoiron/sqlx"
)

const getProducts = `SELECT
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

const getProductsActive = `SELECT
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
    deleted_at,
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

const getProductsTrashed = `SELECT
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
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM products as p
WHERE
    deleted_at IS NOT NULL
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

const getProductsByMerchant = `WITH
    filtered_products AS (
        SELECT
            p.product_id,
            p.merchant_id,
            p.category_id,
            p.weight,
            p.rating,
            p.slug_product,
            p.name,
            p.description,
            p.price,
            p.count_in_stock,
            p.brand,
            p.image_product,
            p.created_at,
            p.updated_at,
            ''::text AS category_name
        FROM products p
        WHERE
            p.deleted_at IS NULL
            AND p.merchant_id = $1
            AND (
                p.name ILIKE '%' || COALESCE($2, '') || '%'
                OR p.description ILIKE '%' || COALESCE($2, '') || '%'
                OR $2 IS NULL
            )
            AND (
                p.category_id = NULLIF($3, 0)
                OR NULLIF($3, 0) IS NULL
            )
            AND (
                p.price >= COALESCE(NULLIF($4, 0), 0)
                AND p.price <= COALESCE(NULLIF($5, 0), 999999999)
            )
    )
SELECT (
        SELECT COUNT(*)
        FROM filtered_products
    ) AS total_count, fp.*
FROM filtered_products fp
ORDER BY fp.created_at DESC
LIMIT $6
OFFSET
    $7`

const getProductsByCategoryName = `WITH
    filtered_products AS (
        SELECT
            p.product_id,
            p.merchant_id,
            p.category_id,
            p.weight,
            p.rating,
            p.slug_product,
            p.name,
            p.description,
            p.price,
            p.count_in_stock,
            p.brand,
            p.image_product,
            p.created_at,
            p.updated_at,
            COALESCE($1::TEXT, '') AS category_name
        FROM products p
        WHERE
            p.deleted_at IS NULL
            AND (
                $1::TEXT IS NULL
                OR p.category_id = NULLIF($1::TEXT, '')::INT
            )
            AND (
                $2::TEXT IS NULL
                OR p.name ILIKE '%' || $2::TEXT || '%'
                OR p.description ILIKE '%' || $2::TEXT || '%'
            )
            AND (
                (
                    $3::INTEGER IS NULL
                    OR p.price >= $3::INTEGER
                )
                AND (
                    $4::INTEGER IS NULL
                    OR p.price <= $4::INTEGER
                )
            )
    )
SELECT (
        SELECT COUNT(*)
        FROM filtered_products
    ) AS total_count, fp.*
FROM filtered_products fp
ORDER BY fp.created_at DESC
LIMIT $5
OFFSET
    $6`

const getProductByID = `SELECT
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
    updated_at
FROM products
WHERE
    product_id = $1
    AND deleted_at IS NULL`

const getProductByIdTrashed = `SELECT * FROM products WHERE product_id = $1`

type productQueryRepository struct {
	db             *sqlx.DB
	categoryRepo categoryadapter.QueryRepository
}

func NewProductQueryRepository(db *sqlx.DB, categoryRepo categoryadapter.QueryRepository) *productQueryRepository {
	return &productQueryRepository{
		db:             db,
		categoryRepo: categoryRepo,
	}
}

func (r *productQueryRepository) FindAll(ctx context.Context, req *requests.FindAllProduct) ([]*db.GetProductsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetProductsRow
	err := r.db.SelectContext(ctx, &rows, getProducts,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, product_errors.ErrFindAllProducts.WithInternal(err)
	}

	return rows, nil
}

func (r *productQueryRepository) FindActive(ctx context.Context, req *requests.FindAllProduct) ([]*db.GetProductsActiveRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetProductsActiveRow
	err := r.db.SelectContext(ctx, &rows, getProductsActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, product_errors.ErrFindActiveProducts.WithInternal(err)
	}

	return rows, nil
}

func (r *productQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllProduct) ([]*db.GetProductsTrashedRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetProductsTrashedRow
	err := r.db.SelectContext(ctx, &rows, getProductsTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, product_errors.ErrFindTrashedProducts.WithInternal(err)
	}

	return rows, nil
}

func (r *productQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllProductByMerchant) ([]*db.GetProductsByMerchantRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetProductsByMerchantRow
	err := r.db.SelectContext(ctx, &rows, getProductsByMerchant,
		int32(req.MerchantID),
		stringPtr(req.Search),
		int32(req.CategoryID),
		int32(IntPtrToInt(req.MinPrice)),
		int32(IntPtrToInt(req.MaxPrice)),
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, product_errors.ErrFindProductsByMerchant.WithInternal(err)
	}

	return rows, nil
}

func (r *productQueryRepository) FindByCategory(ctx context.Context, req *requests.FindAllProductByCategory) ([]*db.GetProductsByCategoryNameRow, error) {
	offset := (req.Page - 1) * req.PageSize

	// The categories table lives in the category service DB (per-service split),
	// so the category name is resolved to its ID via the category service gRPC
	// before querying products locally.
	categoryID := int32(0)
	if req.CategoryName != "" {
		catRes, err := r.categoryRepo.FindAllSearch(ctx, 1, 1, req.CategoryName)
		if err != nil {
			return nil, product_errors.ErrFindProductsByCategory.WithInternal(err)
		}
		if len(catRes) > 0 {
			categoryID = catRes[0].CategoryID
		}
	}

	var rows []*db.GetProductsByCategoryNameRow
	err := r.db.SelectContext(ctx, &rows, getProductsByCategoryName,
		strconv.FormatInt(int64(categoryID), 10),
		req.Search,
		int32(IntPtrToInt(req.MinPrice)),
		int32(IntPtrToInt(req.MaxPrice)),
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, product_errors.ErrFindProductsByCategory.WithInternal(err)
	}

	return rows, nil
}

func (r *productQueryRepository) FindByID(ctx context.Context, product_id int) (*db.GetProductByIDRow, error) {
	var row db.GetProductByIDRow
	err := r.db.GetContext(ctx, &row, getProductByID, int32(product_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound.WithInternal(err)
		}
		return nil, product_errors.ErrProductInternal.WithInternal(err)
	}

	return &row, nil
}

func (r *productQueryRepository) FindByIDTrashed(ctx context.Context, product_id int) (*db.Product, error) {
	var row db.Product
	err := r.db.GetContext(ctx, &row, getProductByIdTrashed, int32(product_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, product_errors.ErrProductNotFound.WithInternal(err)
		}
		return nil, product_errors.ErrProductInternal.WithInternal(err)
	}

	return &row, nil
}
