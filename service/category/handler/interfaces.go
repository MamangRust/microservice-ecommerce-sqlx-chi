package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CategoryQueryHandler interface {
	pb_category.CategoryQueryServiceServer
}

type CategoryCommandHandler interface {
	pb_category.CategoryCommandServiceServer
}

type CategoryHandleGrpc interface {
	FindAll(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategory, error)
	FindById(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategory, error)
	FindByActive(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategoryDeleteAt, error)
	FindByTrashed(ctx context.Context, request *pb_category.FindAllCategoryRequest) (*pb_category.ApiResponsePaginationCategoryDeleteAt, error)
	Create(ctx context.Context, request *pb_category.CreateCategoryRequest) (*pb_category.ApiResponseCategory, error)
	Update(ctx context.Context, request *pb_category.UpdateCategoryRequest) (*pb_category.ApiResponseCategory, error)
	TrashedCategory(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDeleteAt, error)
	RestoreCategory(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDeleteAt, error)
	DeleteCategoryPermanent(ctx context.Context, request *pb_category.FindByIdCategoryRequest) (*pb_category.ApiResponseCategoryDelete, error)
	RestoreAllCategory(ctx context.Context, _ *emptypb.Empty) (*pb_category.ApiResponseCategoryAll, error)
	DeleteAllCategoryPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_category.ApiResponseCategoryAll, error)
}
