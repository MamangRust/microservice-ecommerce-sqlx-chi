package category

import (
	"context"
	"time"

	pbcategory "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/category_errors"
)

// QueryRepository is the contract consumers depend on for category reads.
// It is implemented by grpcQueryAdapter and can be satisfied by a test double.
type QueryRepository interface {
	FindByID(ctx context.Context, categoryID int) (*models.Category, error)
	FindAllSearch(ctx context.Context, page, pageSize int, search string) ([]*models.Category, error)
}

// BulkRepository enumerates categories page by page for the stats backfill, which
// needs every category's name to denormalize order items.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Category, int, error)
}

type grpcQueryAdapter struct {
	client pbcategory.CategoryQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewQueryAdapter wraps a generated category query client into a QueryRepository.
// Zero or more guard options attach a per-call timeout, bulkhead and circuit
// breaker around every outbound gRPC call; with no options the guard stays nil
// and calls pass straight through.
func NewQueryAdapter(client pbcategory.CategoryQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &grpcQueryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// NewAdapter is an alias for NewQueryAdapter kept for callers that build the
// category adapter alongside command-capable adapters.
func NewAdapter(client pbcategory.CategoryQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	return NewQueryAdapter(client, opts...)
}

// NewBulkAdapter wraps a generated category query client for bulk enumeration (the
// stats backfill).
func NewBulkAdapter(client pbcategory.CategoryQueryServiceClient, opts ...adapter.GuardOption) BulkRepository {
	a := &grpcQueryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *grpcQueryAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *grpcQueryAdapter) FindByID(ctx context.Context, categoryID int) (*models.Category, error) {
	var res *pbcategory.ApiResponseCategory
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindById(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(categoryID)})
		return callErr
	})
	if err != nil {
		return nil, category_errors.ErrCategoryNotFound.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *grpcQueryAdapter) FindAllSearch(ctx context.Context, page, pageSize int, search string) ([]*models.Category, error) {
	var res *pbcategory.ApiResponsePaginationCategory
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindAll(ctx, &pbcategory.FindAllCategoryRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
			Search:   search,
		})
		return callErr
	})
	if err != nil {
		return nil, category_errors.ErrFindAllCategory.WithInternal(err)
	}
	return toModels(res.Data), nil
}

// FindAll walks every category page by page for the backfill. The returned int is
// the total record count reported by the service.
func (a *grpcQueryAdapter) FindAll(ctx context.Context, page, pageSize int) ([]*models.Category, int, error) {
	var res *pbcategory.ApiResponsePaginationCategory
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindAll(ctx, &pbcategory.FindAllCategoryRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil {
		return nil, 0, category_errors.ErrFindAllCategory.WithInternal(err)
	}
	categories := toModels(res.Data)
	total := len(categories)
	if res.Pagination != nil {
		total = int(res.Pagination.TotalRecords)
	}
	return categories, total, nil
}

func toModel(c *pbcategory.CategoryResponse) *models.Category {
	if c == nil {
		return nil
	}
	return &models.Category{
		CategoryID:    c.Id,
		Name:          c.Name,
		Description:   &c.Description,
		SlugCategory:  &c.SlugCategory,
		ImageCategory: &c.ImageCategory,
		CreatedAt:     parseTimeString(c.CreatedAt),
		UpdatedAt:     parseTimeString(c.UpdatedAt),
	}
}

func toModels(cs []*pbcategory.CategoryResponse) []*models.Category {
	out := make([]*models.Category, 0, len(cs))
	for _, c := range cs {
		out = append(out, toModel(c))
	}
	return out
}

func parseTimeString(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05.000", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
		if err != nil {
			return nil
		}
	}
	return &t
}
