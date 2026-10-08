package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/database/schema"
	merchantadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/merchant"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

type MerchantAwardQueryRepository interface {
	FindAll(
		ctx context.Context,
		req *requests.FindAllMerchant,
	) ([]*db.GetMerchantCertificationsAndAwardsRow, error)

	FindActive(
		ctx context.Context,
		req *requests.FindAllMerchant,
	) ([]*db.GetMerchantCertificationsAndAwardsActiveRow, error)

	FindTrashed(
		ctx context.Context,
		req *requests.FindAllMerchant,
	) ([]*db.GetMerchantCertificationsAndAwardsTrashedRow, error)

	FindByID(
		ctx context.Context,
		user_id int,
	) (*db.GetMerchantCertificationOrAwardRow, error)
}

type MerchantAwardCommandRepository interface {
	Create(
		ctx context.Context,
		request *requests.CreateMerchantCertificationOrAwardRequest,
	) (*db.CreateMerchantCertificationOrAwardRow, error)

	Update(ctx context.Context, request *requests.UpdateMerchantCertificationOrAwardRequest) (*db.UpdateMerchantCertificationOrAwardRow, error)

	Trash(
		ctx context.Context,
		merchant_id int,
	) (*db.MerchantCertificationsAndAward, error)

	Restore(
		ctx context.Context,
		merchant_id int,
	) (*db.MerchantCertificationsAndAward, error)

	DeletePermanent(
		ctx context.Context,
		merchant_id int,
	) (bool, error)

	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}

type MerchantQueryRepository = merchantadapter.QueryRepository
