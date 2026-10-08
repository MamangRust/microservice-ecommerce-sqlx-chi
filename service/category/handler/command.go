package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-category/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	category_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/category_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type categoryCommandHandler struct {
	pb_category.UnimplementedCategoryCommandServiceServer
	service service.CategoryCommandService
	logger  logger.LoggerInterface
}

func NewCategoryCommandHandler(service service.CategoryCommandService, logger logger.LoggerInterface) pb_category.CategoryCommandServiceServer {
	return &categoryCommandHandler{
		service: service,
		logger:  logger,
	}
}

func (h *categoryCommandHandler) Create(ctx context.Context, request *pb_category.CreateCategoryRequest) (*pb_category.ApiResponseCategory, error) {
	slug := request.GetSlugCategory()
	req := &requests.CreateCategoryRequest{
		Name:          request.GetName(),
		Description:   request.GetDescription(),
		SlugCategory:  &slug,
		ImageCategory: request.GetImageCategory(),
	}

	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateCreateCategory
	}

	category, err := h.service.Create(ctx, req)
	if err != nil {
		return nil, category_errors.ErrGrpcCreateCategory
	}

	return &pb_category.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully created category",
		Data:    (&Handler{}).mapToCategoryResponse(category).(*pb_category.CategoryResponse),
	}, nil
}

func (h *categoryCommandHandler) Update(ctx context.Context, request *pb_category.UpdateCategoryRequest) (*pb_category.ApiResponseCategory, error) {
	id := int(request.GetCategoryId())

	if id == 0 {
		return nil, category_errors.ErrGrpcCategoryInvalidId
	}

	slug := request.GetSlugCategory()
	req := &requests.UpdateCategoryRequest{
		CategoryID:    &id,
		Name:          request.GetName(),
		Description:   request.GetDescription(),
		SlugCategory:  &slug,
		ImageCategory: request.GetImageCategory(),
	}

	if err := req.Validate(); err != nil {
		return nil, category_errors.ErrGrpcValidateUpdateCategory
	}

	category, err := h.service.Update(ctx, req)
	if err != nil {
		return nil, category_errors.ErrGrpcUpdateCategory
	}

	return &pb_category.ApiResponseCategory{
		Status:  "success",
		Message: "Successfully updated category",
		Data:    (&Handler{}).mapToCategoryResponse(category).(*pb_category.CategoryResponse),
	}, nil
}

func (h *categoryCommandHandler) TrashedCategory(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, category_errors.ErrGrpcCategoryInvalidId
	}

	category, err := h.service.Trash(ctx, id)
	if err != nil {
		return nil, category_errors.ErrGrpcCategoryNotFound
	}

	return &pb_category.ApiResponseCategoryDeleteAt{
		Status:  "success",
		Message: "Successfully trashed category",
		Data:    (&Handler{}).mapToCategoryResponse(category).(*pb_category.CategoryResponseDeleteAt),
	}, nil
}

func (h *categoryCommandHandler) RestoreCategory(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, category_errors.ErrGrpcCategoryInvalidId
	}

	category, err := h.service.Restore(ctx, id)
	if err != nil {
		return nil, category_errors.ErrGrpcCategoryNotFound
	}

	return &pb_category.ApiResponseCategoryDeleteAt{
		Status:  "success",
		Message: "Successfully restored category",
		Data:    (&Handler{}).mapToCategoryResponse(category).(*pb_category.CategoryResponseDeleteAt),
	}, nil
}

func (h *categoryCommandHandler) DeleteCategoryPermanent(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDelete, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, category_errors.ErrGrpcCategoryInvalidId
	}

	_, err := h.service.DeletePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_category.ApiResponseCategoryDelete{
		Status:  "success",
		Message: "Successfully deleted category permanently",
	}, nil
}

func (h *categoryCommandHandler) RestoreAllCategory(ctx context.Context, _ *emptypb.Empty) (*pb_category.ApiResponseCategoryAll, error) {
	_, err := h.service.RestoreAll(ctx)
	if err != nil {
		return nil, category_errors.ErrGrpcCategoryNotFound
	}

	return &pb_category.ApiResponseCategoryAll{
		Status:  "success",
		Message: "Successfully restored all categories",
	}, nil
}

func (h *categoryCommandHandler) DeleteAllCategoryPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_category.ApiResponseCategoryAll, error) {
	_, err := h.service.DeleteAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_category.ApiResponseCategoryAll{
		Status:  "success",
		Message: "Successfully deleted all categories permanently",
	}, nil
}
