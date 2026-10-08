package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-ecommerce-grpc-category/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	category_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/category_errors"
)

type categoryQueryHandler struct {
	pb_category.UnimplementedCategoryQueryServiceServer
	service service.CategoryQueryService
	logger  logger.LoggerInterface
}

func NewCategoryQueryHandler(service service.CategoryQueryService, logger logger.LoggerInterface) pb_category.CategoryQueryServiceServer {
	return &categoryQueryHandler{
		service: service,
		logger:  logger,
	}
}

func (h *categoryQueryHandler) FindAll(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategory, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	categories, totalRecords, err := h.service.FindAll(ctx, &reqService)
	if err != nil {
		return nil, category_errors.ErrGrpcFindAllCategory
	}

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(math.Ceil(float64(*totalRecords) / float64(pageSize))),
		TotalRecords: int32(*totalRecords),
	}

	results := make([]*pb_category.CategoryResponse, len(categories))
	for i, v := range categories {
		results[i] = (&Handler{}).mapToCategoryResponse(v).(*pb_category.CategoryResponse)
	}

	return &pb_category.ApiResponsePaginationCategory{
		Status:     "success",
		Message:    "Successfully fetched categories",
		Data:       results,
		Pagination: paginationMeta,
	}, nil
}

func (h *categoryQueryHandler) FindById(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategory, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, category_errors.ErrGrpcCategoryInvalidId
	}

	category, err := h.service.FindByID(ctx, id)
	if err != nil {
		return nil, category_errors.ErrGrpcCategoryNotFound
	}

	return &pb_category.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully fetched category",
		Data:    (&Handler{}).mapToCategoryResponse(category).(*pb_category.CategoryResponse),
	}, nil
}

func (h *categoryQueryHandler) FindByActive(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategoryDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	categories, totalRecords, err := h.service.FindActive(ctx, &reqService)
	if err != nil {
		return nil, category_errors.ErrGrpcFindAllCategory
	}

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(math.Ceil(float64(*totalRecords) / float64(pageSize))),
		TotalRecords: int32(*totalRecords),
	}

	results := make([]*pb_category.CategoryResponseDeleteAt, len(categories))
	for i, v := range categories {
		results[i] = (&Handler{}).mapToCategoryResponse(v).(*pb_category.CategoryResponseDeleteAt)
	}

	return &pb_category.ApiResponsePaginationCategoryDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active categories",
		Data:       results,
		Pagination: paginationMeta,
	}, nil
}

func (h *categoryQueryHandler) FindByTrashed(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategoryDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCategory{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	categories, totalRecords, err := h.service.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, category_errors.ErrGrpcFindAllCategory
	}

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(math.Ceil(float64(*totalRecords) / float64(pageSize))),
		TotalRecords: int32(*totalRecords),
	}

	results := make([]*pb_category.CategoryResponseDeleteAt, len(categories))
	for i, v := range categories {
		results[i] = (&Handler{}).mapToCategoryResponse(v).(*pb_category.CategoryResponseDeleteAt)
	}

	return &pb_category.ApiResponsePaginationCategoryDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed categories",
		Data:       results,
		Pagination: paginationMeta,
	}, nil
}
