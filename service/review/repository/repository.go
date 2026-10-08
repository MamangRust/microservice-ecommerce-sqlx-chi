package repository

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

// GuardOptions collects the dependency guards applied to the product/user
// adapters built inside NewRepositories. Product and user reads get separate
// breakers so one failing dependency cannot starve the other.
type GuardOptions struct {
	User    []adapter.GuardOption
	Product []adapter.GuardOption
}

type Repositories struct {
	ProductQuery  ProductQueryRepository
	ReviewQuery   ReviewQueryRepository
	UserQuery     UserQueryRepository
	ReviewCommand ReviewCommandRepository
}

// NewRepositories assembles review's own Mongo repositories and builds the
// shared product/user gRPC adapters from the raw clients under their guards.
func NewRepositories(DB *sqlx.DB,
	userQueryClient pb_user.UserQueryServiceClient,
	productQueryClient pb_product.ProductQueryServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	productQuery := productadapter.NewQueryAdapter(productQueryClient, g.Product...)

	return &Repositories{
		ProductQuery:  productQuery,
		ReviewQuery:   NewReviewQueryRepository(DB, productQuery),
		UserQuery:     useradapter.NewQueryAdapter(userQueryClient, g.User...),
		ReviewCommand: NewReviewCommandRepository(DB),
	}
}
