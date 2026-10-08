package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-cart/database/schema"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

type CartQueryRepository interface {
	FindCarts(
		ctx context.Context,
		req *requests.FindAllCarts,
	) ([]*db.GetCartsRow, error)
}

type CartCommandRepository interface {
	CreateCart(
		ctx context.Context,
		req *requests.CartCreateRecord,
	) (*db.CreateCartRow, error)

	DeletePermanent(
		ctx context.Context,
		req *requests.DeleteCartRequest,
	) (bool, error)

	DeleteAllPermanently(
		ctx context.Context,
		req *requests.DeleteAllCartRequest,
	) (bool, error)
}

type ProductQueryRepository = productadapter.QueryRepository

type UserQueryRepository = useradapter.QueryRepository
