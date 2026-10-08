package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-cart/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/cart_errors"
	"github.com/jmoiron/sqlx"
)

const createCart = `INSERT INTO
    "carts" (
        "user_id",
        "product_id",
        "name",
        "price",
        "image",
        "quantity",
        "weight"
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    cart_id,
    user_id,
    product_id,
    name,
    price,
    image,
    quantity,
    weight,
    created_at,
    updated_at`

const deleteCartByIdAndUserId = `DELETE FROM "carts" WHERE "cart_id" = $1 AND "user_id" = $2`

const deleteAllCartByUserId = `DELETE FROM "carts"
WHERE
    "cart_id" = ANY ($1::int[])
    AND "user_id" = $2`

type cartCommandRepository struct {
	db *sqlx.DB
}

func NewCartCommandRepository(db *sqlx.DB) CartCommandRepository {
	return &cartCommandRepository{
		db: db,
	}
}

func (r *cartCommandRepository) CreateCart(ctx context.Context, req *requests.CartCreateRecord) (*db.CreateCartRow, error) {
	var row db.CreateCartRow

	if err := r.db.GetContext(ctx, &row, createCart,
		int32(req.UserID),
		int32(req.ProductID),
		req.Name,
		int32(req.Price),
		req.ImageProduct,
		int32(req.Quantity),
		int32(req.Weight),
	); err != nil {
		return nil, cart_errors.ErrCreateCart.WithInternal(err)
	}

	return &row, nil
}

func (r *cartCommandRepository) DeletePermanent(ctx context.Context, req *requests.DeleteCartRequest) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteCartByIdAndUserId,
		int32(req.CartID),
		int32(req.UserID),
	); err != nil {
		return false, cart_errors.ErrDeleteCartPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *cartCommandRepository) DeleteAllPermanently(ctx context.Context, req *requests.DeleteAllCartRequest) (bool, error) {
	cartIDs := make([]int32, len(req.CartIds))

	for i, id := range req.CartIds {
		cartIDs[i] = int32(id)
	}

	if _, err := r.db.ExecContext(ctx, deleteAllCartByUserId,
		cartIDs,
		int32(req.UserID),
	); err != nil {
		return false, cart_errors.ErrDeleteAllCarts.WithInternal(err)
	}

	return true, nil
}
