package repository

import (
	pb_user "github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type GuardOptions struct {
	User []adapter.GuardOption
}

type Repositories struct {
	MerchantQuery           MerchantQueryRepository
	MerchantCommand         MerchantCommandRepository
	MerchantDocumentCommand MerchantDocumentCommandRepository
	MerchantDocumentQuery   MerchantDocumentQueryRepository
	UserQuery               useradapter.QueryRepository
}

func NewRepositories(DB *sqlx.DB, userQueryClient pb_user.UserQueryServiceClient, guards ...GuardOptions) *Repositories {
	var g GuardOptions

	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		MerchantQuery:           NewMerchantQueryRepository(DB),
		MerchantCommand:         NewMerchantCommandRepository(DB),
		MerchantDocumentCommand: NewMerchantDocumentCommandRepository(DB),
		MerchantDocumentQuery:   NewMerchantDocumentQueryRepository(DB),
		UserQuery:               useradapter.NewQueryAdapter(userQueryClient, g.User...),
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
