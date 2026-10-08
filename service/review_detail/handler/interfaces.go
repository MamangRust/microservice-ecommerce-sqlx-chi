package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review_detail"
)

type ReviewDetailQueryHandler interface {
	pb_review_detail.ReviewDetailQueryServiceServer
}

type ReviewDetailCommandHandler interface {
	pb_review_detail.ReviewDetailCommandServiceServer
}

type ReviewDetailHandleGrpc interface {
	FindAll(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review_detail.ApiResponsePaginationReviewDetails, error)
	FindById(ctx context.Context, request *pb_review_detail.FindByIdReviewDetailRequest) (*pb_review_detail.ApiResponseReviewDetail, error)
	FindByActive(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review_detail.ApiResponsePaginationReviewDetailsDeleteAt, error)
	FindByTrashed(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review_detail.ApiResponsePaginationReviewDetailsDeleteAt, error)
	Create(ctx context.Context, request *pb_review_detail.CreateReviewDetailRequest) (*pb_review_detail.ApiResponseReviewDetail, error)
	Update(ctx context.Context, request *pb_review_detail.UpdateReviewDetailRequest) (*pb_review_detail.ApiResponseReviewDetail, error)
	TrashedReviewDetail(ctx context.Context, request *pb_review_detail.FindByIdReviewDetailRequest) (*pb_review_detail.ApiResponseReviewDetailDeleteAt, error)
	RestoreReviewDetail(ctx context.Context, request *pb_review_detail.FindByIdReviewDetailRequest) (*pb_review_detail.ApiResponseReviewDetailDeleteAt, error)
	DeleteReviewDetailPermanent(ctx context.Context, request *pb_review_detail.FindByIdReviewDetailRequest) (*pb_review.ApiResponseReviewDelete, error)
}
