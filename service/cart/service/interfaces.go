package service

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-cart/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

type CartQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCarts) ([]*db.GetCartsRow, *int, error)
}

type CartCommandService interface {
	Create(ctx context.Context, req *requests.CreateCartRequest) (*db.Cart, error)
	DeletePermanent(ctx context.Context, req *requests.DeleteCartRequest) (bool, error)
	DeleteAll(ctx context.Context, req *requests.DeleteAllCartRequest) (bool, error)
}
