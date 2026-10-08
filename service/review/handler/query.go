package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/review"
	"github.com/MamangRust/microservice-ecommerce-grpc-review/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
)

type reviewQueryHandler struct {
	pb_review.UnimplementedReviewQueryServiceServer
	reviewService service.ReviewQueryService
	logger        logger.LoggerInterface
}

func NewReviewQueryHandler(reviewService service.ReviewQueryService, logger logger.LoggerInterface) pb_review.ReviewQueryServiceServer {
	return &reviewQueryHandler{
		reviewService: reviewService,
		logger:        logger,
	}
}

func (h *reviewQueryHandler) FindAll(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review.ApiResponsePaginationReview, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	reviews, totalRecords, err := h.reviewService.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReviews := hMapping.mapResponse(reviews).([]*pb_review.ReviewResponse)

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_review.ApiResponsePaginationReview{
		Status:     "success",
		Message:    "Successfully fetched reviews",
		Data:       protoReviews,
		Pagination: paginationMeta,
	}, nil
}

func (h *reviewQueryHandler) FindByProduct(ctx context.Context, request *pb_review.FindAllReviewProductRequest) (*pb_review.ApiResponsePaginationReviewDetail, error) {
	product_id := int(request.GetProductId())
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllReviewByProduct{
		ProductID: product_id,
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
	}

	reviews, totalRecords, err := h.reviewService.FindByProduct(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReviews := hMapping.mapResponse(reviews).([]*pb_review.ReviewsDetailResponse)

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_review.ApiResponsePaginationReviewDetail{
		Status:     "success",
		Message:    "Successfully fetched product reviews",
		Data:       protoReviews,
		Pagination: paginationMeta,
	}, nil
}

func (h *reviewQueryHandler) FindByMerchant(ctx context.Context, request *pb_review.FindAllReviewMerchantRequest) (*pb_review.ApiResponsePaginationReviewDetail, error) {
	merchant_id := int(request.GetMerchantId())
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllReviewByMerchant{
		MerchantID: merchant_id,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
	}

	reviews, totalRecords, err := h.reviewService.FindByMerchant(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReviews := hMapping.mapResponse(reviews).([]*pb_review.ReviewsDetailResponse)

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))
	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_review.ApiResponsePaginationReviewDetail{
		Status:     "success",
		Message:    "Successfully fetched merchant reviews",
		Data:       protoReviews,
		Pagination: paginationMeta,
	}, nil
}

func (h *reviewQueryHandler) FindByActive(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review.ApiResponsePaginationReviewDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	reviews, totalRecords, err := h.reviewService.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReviews := hMapping.mapResponse(reviews).([]*pb_review.ReviewResponseDeleteAt)

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_review.ApiResponsePaginationReviewDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active reviews",
		Data:       protoReviews,
		Pagination: paginationMeta,
	}, nil
}

func (h *reviewQueryHandler) FindByTrashed(ctx context.Context, request *pb_review.FindAllReviewRequest) (*pb_review.ApiResponsePaginationReviewDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllReview{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	reviews, totalRecords, err := h.reviewService.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	var hMapping reviewHandleGrpc
	protoReviews := hMapping.mapResponse(reviews).([]*pb_review.ReviewResponseDeleteAt)

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_review.ApiResponsePaginationReviewDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed reviews",
		Data:       protoReviews,
		Pagination: paginationMeta,
	}, nil
}
