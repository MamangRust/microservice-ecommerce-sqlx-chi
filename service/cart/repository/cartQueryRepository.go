package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-cart/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/cart_errors"
	"github.com/jmoiron/sqlx"
)

const getCarts = `SELECT
    cart_id,
    user_id,
    product_id,
    name,
    price,
    image,
    quantity,
    weight,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM carts
WHERE
    deleted_at IS NULL
    AND user_id = $1
    AND (
        $2::TEXT IS NULL
        OR name ILIKE '%' || $2 || '%'
        OR price::TEXT ILIKE '%' || $2 || '%'
    )
ORDER BY created_at DESC
LIMIT $3
OFFSET
    $4`

type cartQueryRepository struct {
	db *sqlx.DB
}

func NewCartQueryRepository(db *sqlx.DB) CartQueryRepository {
	return &cartQueryRepository{
		db: db,
	}
}

func (r *cartQueryRepository) FindCarts(ctx context.Context, req *requests.FindAllCarts) ([]*db.GetCartsRow, error) {
	offset := (req.Page - 1) * req.PageSize

	var rows []*db.GetCartsRow

	if err := r.db.SelectContext(ctx, &rows, getCarts,
		int32(req.UserID),
		req.Search,
		int32(req.PageSize),
		int32(offset),
	); err != nil {
		return nil, cart_errors.ErrFindAllCarts.WithInternal(err)
	}

	return rows, nil
}
