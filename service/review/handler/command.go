package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	review_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/review"
	"google.golang.org/protobuf/types/known/emptypb"
)

type reviewCommandHandler struct {
	pb_review.UnimplementedReviewCommandServiceServer
	reviewService service.ReviewCommandService
	logger        logger.LoggerInterface
}

func NewReviewCommandHandler(reviewService service.ReviewCommandService, logger logger.LoggerInterface) pb_review.ReviewCommandServiceServer {
	return &reviewCommandHandler{
		reviewService: reviewService,
		logger:        logger,
	}
}

func (h *reviewCommandHandler) Create(ctx context.Context, request *pb_review.CreateReviewRequest) (*pb_review.ApiResponseReview, error) {
	req := &requests.CreateReviewRequest{
		UserID:    int(request.GetUserId()),
		ProductID: int(request.GetProductId()),
		Rating:    int(request.GetRating()),
		Comment:   request.GetComment(),
	}

	if err := req.Validate(); err != nil {
		return nil, review_errors.ErrGrpcValidateCreateReview
	}

	review, err := h.reviewService.Create(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReview := hMapping.mapResponse(review).(*pb_review.ReviewResponse)

	return &pb_review.ApiResponseReview{
		Status:  "success",
		Message: "Successfully created review",
		Data:    protoReview,
	}, nil
}

func (h *reviewCommandHandler) Update(ctx context.Context, request *pb_review.UpdateReviewRequest) (*pb_review.ApiResponseReview, error) {
	id := int(request.GetReviewId())

	if id == 0 {
		return nil, review_errors.ErrGrpcInvalidID
	}

	req := &requests.UpdateReviewRequest{
		ReviewID: &id,
		Name:     request.GetName(),
		Rating:   int(request.GetRating()),
		Comment:  request.GetComment(),
	}

	if err := req.Validate(); err != nil {
		return nil, review_errors.ErrGrpcValidateUpdateReview
	}

	review, err := h.reviewService.Update(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReview := hMapping.mapResponse(review).(*pb_review.ReviewResponse)

	return &pb_review.ApiResponseReview{
		Status:  "success",
		Message: "Successfully updated review",
		Data:    protoReview,
	}, nil
}

func (h *reviewCommandHandler) TrashedReview(ctx context.Context, request *pb_review.FindByIdReviewRequest) (*pb_review.ApiResponseReviewDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, review_errors.ErrGrpcInvalidID
	}

	review, err := h.reviewService.Trash(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReview := hMapping.mapResponse(review).(*pb_review.ReviewResponseDeleteAt)

	return &pb_review.ApiResponseReviewDeleteAt{
		Status:  "success",
		Message: "Successfully trashed review",
		Data:    protoReview,
	}, nil
}

func (h *reviewCommandHandler) RestoreReview(ctx context.Context, request *pb_review.FindByIdReviewRequest) (*pb_review.ApiResponseReviewDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, review_errors.ErrGrpcInvalidID
	}

	review, err := h.reviewService.Restore(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReview := hMapping.mapResponse(review).(*pb_review.ReviewResponseDeleteAt)

	return &pb_review.ApiResponseReviewDeleteAt{
		Status:  "success",
		Message: "Successfully restored review",
		Data:    protoReview,
	}, nil
}

func (h *reviewCommandHandler) DeleteReviewPermanent(ctx context.Context, request *pb_review.FindByIdReviewRequest) (*pb_review.ApiResponseReviewDelete, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, review_errors.ErrGrpcInvalidID
	}

	_, err := h.reviewService.DeletePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_review.ApiResponseReviewDelete{
		Status:  "success",
		Message: "Successfully deleted review permanently",
	}, nil
}

func (h *reviewCommandHandler) RestoreAllReview(ctx context.Context, _ *emptypb.Empty) (*pb_review.ApiResponseReviewAll, error) {
	_, err := h.reviewService.RestoreAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_review.ApiResponseReviewAll{
		Status:  "success",
		Message: "Successfully restored all reviews",
	}, nil
}

func (h *reviewCommandHandler) DeleteAllReviewPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_review.ApiResponseReviewAll, error) {
	_, err := h.reviewService.DeleteAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_review.ApiResponseReviewAll{
		Status:  "success",
		Message: "Successfully deleted all reviews permanently",
	}, nil
}
