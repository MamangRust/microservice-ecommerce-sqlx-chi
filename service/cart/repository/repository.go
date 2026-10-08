package repository

import (
	pb_product "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	User    []adapter.GuardOption
	Product []adapter.GuardOption
}

type Repositories struct {
	CartQuery    CartQueryRepository
	CartCommand  CartCommandRepository
	UserQuery    useradapter.QueryRepository
	ProductQuery productadapter.QueryRepository
}

func NewRepositories(DB *sqlx.DB,
	userQueryClient pb_user.UserQueryServiceClient,
	productQueryClient pb_product.ProductQueryServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		CartQuery:    NewCartQueryRepository(DB),
		CartCommand:  NewCartCommandRepository(DB),
		UserQuery:    useradapter.NewQueryAdapter(userQueryClient, g.User...),
		ProductQuery: productadapter.NewQueryAdapter(productQueryClient, g.Product...),
	}
}
