package reviewapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type ReviewBaseResponseMapper interface {
	ToResponseReview(pbResponse *pb_review.ReviewResponse) *response.ReviewResponse
	ToResponsesReview(pbResponses []*pb_review.ReviewResponse) []*response.ReviewResponse
	ToResponseReviewsDetail(pbResponse *pb_review.ReviewsDetailResponse) *response.ReviewsDetailResponse
	ToResponsesReviewsDetail(pbResponses []*pb_review.ReviewsDetailResponse) []*response.ReviewsDetailResponse
}

type ReviewQueryResponseMapper interface {
	ReviewBaseResponseMapper
	ToApiResponseReview(pbResponse *pb_review.ApiResponseReview) *response.ApiResponseReview
	ToApiResponsesReview(pbResponse *pb_review.ApiResponsesReview) *response.ApiResponsesReview
	ToApiResponsePaginationReview(pbResponse *pb_review.ApiResponsePaginationReview) *response.ApiResponsePaginationReview
	ToApiResponsePaginationReviewsDetail(pbResponse *pb_review.ApiResponsePaginationReviewDetail) *response.ApiResponsePaginationReviewsDetail
	ToApiResponsePaginationReviewDeleteAt(pbResponse *pb_review.ApiResponsePaginationReviewDeleteAt) *response.ApiResponsePaginationReviewDeleteAt
}

type ReviewCommandResponseMapper interface {
	ReviewBaseResponseMapper
	ToResponseReviewDeleteAt(pbResponse *pb_review.ReviewResponseDeleteAt) *response.ReviewResponseDeleteAt
	ToResponsesReviewDeleteAt(pbResponses []*pb_review.ReviewResponseDeleteAt) []*response.ReviewResponseDeleteAt
	ToApiResponseReviewDeleteAt(pbResponse *pb_review.ApiResponseReviewDeleteAt) *response.ApiResponseReviewDeleteAt
	ToApiResponseReviewDelete(pbResponse *pb_review.ApiResponseReviewDelete) *response.ApiResponseReviewDelete
	ToApiResponseReviewAll(pbResponse *pb_review.ApiResponseReviewAll) *response.ApiResponseReviewAll
	ToApiResponsePaginationReviewDeleteAt(pbResponse *pb_review.ApiResponsePaginationReviewDeleteAt) *response.ApiResponsePaginationReviewDeleteAt
}
