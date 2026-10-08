package repository

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-review/database/schema"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	useradapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/user"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

// UserQueryRepository is satisfied by the shared user adapter.
type UserQueryRepository = useradapter.QueryRepository

// ProductQueryRepository is satisfied by the shared product adapter.
type ProductQueryRepository = productadapter.QueryRepository

type ReviewQueryRepository interface {
	FindAll(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsRow, error)
	FindByProduct(ctx context.Context, req *requests.FindAllReviewByProduct) ([]*db.GetReviewByProductIdRow, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllReviewByMerchant) ([]*db.GetReviewByMerchantIdRow, error)
	FindActive(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsActiveRow, error)
	FindTrashed(ctx context.Context, req *requests.FindAllReview) ([]*db.GetReviewsTrashedRow, error)
	FindByID(ctx context.Context, id int) (*db.GetReviewByIDRow, error)
}

type ReviewCommandRepository interface {
	Create(ctx context.Context, request *requests.CreateReviewRequest) (*db.CreateReviewRow, error)
	Update(ctx context.Context, request *requests.UpdateReviewRequest) (*db.UpdateReviewRow, error)
	Trash(ctx context.Context, review_id int) (*db.Review, error)
	Restore(ctx context.Context, review_id int) (*db.Review, error)

	DeletePermanent(
		ctx context.Context,
		review_id int,
	) (bool, error)

	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}
