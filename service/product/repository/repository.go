package repository

import (
	pb_category "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	categoryadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/category"
	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	Category []adapter.GuardOption
	Merchant []adapter.GuardOption
}

type Repositories struct {
	ProductQuery   ProductQueryRepository
	ProductCommand ProductCommandRepository
	CategoryQuery  categoryadapter.QueryRepository
	MerchantQuery  merchantadapter.QueryRepository
}

func NewRepositories(db *sqlx.DB,
	categoryQueryClient pb_category.CategoryQueryServiceClient,
	merchantQueryClient pb_merchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	categoryQuery := categoryadapter.NewAdapter(categoryQueryClient, g.Category...)

	return &Repositories{
		ProductQuery:   NewProductQueryRepository(db, categoryQuery),
		ProductCommand: NewProductCommandRepository(db),
		CategoryQuery:  categoryQuery,
		MerchantQuery:  merchantadapter.NewQueryAdapter(merchantQueryClient, g.Merchant...),
	}
}
