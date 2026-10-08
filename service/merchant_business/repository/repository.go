package repository

import (
	pb_merchant "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	Merchant []adapter.GuardOption
}

type Repositories struct {
	MerchantBusinessQuery   MerchantBusinessQueryRepository
	MerchantBusinessCommand MerchantBusinessCommandRepository
	MerchantQuery           merchantadapter.QueryRepository
}

func NewRepositories(db *sqlx.DB, merchantQueryClient pb_merchant.MerchantQueryServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		MerchantBusinessQuery:   NewMerchantBusinessQueryRepository(db),
		MerchantBusinessCommand: NewMerchantBusinessCommandRepository(db),
		MerchantQuery:           merchantadapter.NewQueryAdapter(merchantQueryClient, g.Merchant...),
	}
}
