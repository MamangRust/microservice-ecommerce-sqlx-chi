package service

import (
	"context"

	db "github.com/MamangRust/microservice-ecommerce-grpc-category/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
)

// F5: the legacy OLTP category stats service (CategoryStatsService,
// CategoryStatsByIdService, CategoryStatsByMerchantService) was removed — stats
// are served by service/stats_reader from ClickHouse.

type CategoryQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, *int, error)
	FindActive(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, *int, error)
	FindTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, *int, error)

	FindByID(ctx context.Context, categoryID int) (*db.GetCategoryByIDRow, error)
}

type CategoryCommandService interface {
	Create(ctx context.Context, req *requests.CreateCategoryRequest) (*db.CreateCategoryRow, error)
	Update(ctx context.Context, req *requests.UpdateCategoryRequest) (*db.UpdateCategoryRow, error)

	Trash(ctx context.Context, categoryID int) (*db.Category, error)
	Restore(ctx context.Context, categoryID int) (*db.Category, error)
	DeletePermanent(ctx context.Context, categoryID int) (bool, error)

	RestoreAll(ctx context.Context) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}
