package product

import (
	"context"
	"time"

	pbproduct "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	product_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/product_errors"
)

// QueryRepository is the contract consumers depend on for product reads.
type QueryRepository interface {
	FindByID(ctx context.Context, productID int) (*models.Product, error)
	// FindIDsByMerchant returns just the product IDs owned by a merchant, for
	// consumers (review) that only need the IDs to filter their own tables.
	FindIDsByMerchant(ctx context.Context, merchantID int) ([]int32, error)
}

// CommandRepository is the contract consumers depend on for product writes
// (stock adjustments). CreateProduct/UpdateProduct live entirely inside the
// owning service and are not exposed here.
type CommandRepository interface {
	UpdateProductCountStock(ctx context.Context, productID int, stock int) (*models.Product, error)
	AdjustProductStock(ctx context.Context, productID int, delta int, operationID string) (*models.Product, error)
}

// BulkRepository enumerates products page by page for the stats backfill, which
// needs every product's category to denormalize order items.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Product, int, error)
}

// Repository implements both QueryRepository and CommandRepository on top of the
// generated product query/command clients.
type Repository struct {
	query   pbproduct.ProductQueryServiceClient
	command pbproduct.ProductCommandServiceClient
	guard   *resilience.DependencyGuard
}

// NewAdapter wraps the generated product clients. Zero or more guard options
// attach a per-call timeout, bulkhead and circuit breaker around every outbound
// gRPC call; with no options the guard stays nil and calls pass straight
// through.
func NewAdapter(query pbproduct.ProductQueryServiceClient, command pbproduct.ProductCommandServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{query: query, command: command}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewQueryAdapter returns the adapter restricted to the QueryRepository surface
// for consumers that only read (cart, review).
func NewQueryAdapter(query pbproduct.ProductQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	r := &Repository{query: query}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewCommandAdapter returns the adapter restricted to the CommandRepository surface
// for consumers that only write (currently none — included for symmetry).
func NewCommandAdapter(command pbproduct.ProductCommandServiceClient, opts ...adapter.GuardOption) CommandRepository {
	r := &Repository{command: command}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (a *Repository) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *Repository) FindByID(ctx context.Context, productID int) (*models.Product, error) {
	var res *pbproduct.ApiResponseProduct
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.query.FindById(ctx, &pbproduct.FindByIdProductRequest{Id: int32(productID)})
		return callErr
	})
	if err != nil {
		return nil, product_errors.ErrProductNotFound.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) UpdateProductCountStock(ctx context.Context, productID int, stock int) (*models.Product, error) {
	var res *pbproduct.ApiResponseProduct
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.UpdateProductCountStock(ctx, &pbproduct.UpdateProductCountStockRequest{
			ProductId: int32(productID),
			Stock:     int32(stock),
		})
		return callErr
	})
	if err != nil {
		return nil, product_errors.ErrUpdateProductCountStock.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) AdjustProductStock(ctx context.Context, productID int, delta int, operationID string) (*models.Product, error) {
	var res *pbproduct.ApiResponseProduct
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.command.AdjustProductStock(ctx, &pbproduct.AdjustProductStockRequest{
			ProductId:   int32(productID),
			Delta:       int32(delta),
			OperationId: operationID,
		})
		return callErr
	})
	if err != nil {
		return nil, product_errors.ErrUpdateProductCountStock.WithInternal(err)
	}
	return toModel(res.Data), nil
}

func (a *Repository) FindIDsByMerchant(ctx context.Context, merchantID int) ([]int32, error) {
	var res *pbproduct.ApiResponsePaginationProduct
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.query.FindByMerchant(ctx, &pbproduct.FindAllProductMerchantRequest{
			MerchantId: int32(merchantID),
			Page:       1,
			PageSize:   100000,
		})
		return callErr
	})
	if err != nil {
		return nil, product_errors.ErrProductNotFound.WithInternal(err)
	}

	ids := make([]int32, 0, len(res.Data))
	for _, p := range res.Data {
		ids = append(ids, p.Id)
	}
	return ids, nil
}

func toModel(data *pbproduct.ProductResponse) *models.Product {
	if data == nil {
		return nil
	}
	description := data.Description
	weight := data.Weight
	image := data.ImageProduct
	return &models.Product{
		ProductID:    data.Id,
		MerchantID:   data.MerchantId,
		CategoryID:   data.CategoryId,
		Name:         data.Name,
		Description:  &description,
		Price:        data.Price,
		CountInStock: data.CountInStock,
		Weight:       &weight,
		ImageProduct: &image,
		CreatedAt:    parseTime(data.CreatedAt),
		UpdatedAt:    parseTime(data.UpdatedAt),
	}
}

// FindAll walks every product page by page. The backfill uses it to build the
// product_id -> category_id map that denormalizes order items.
func (a *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Product, int, error) {
	var resp *pbproduct.ApiResponsePaginationProduct
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = a.query.FindAll(ctx, &pbproduct.FindAllProductRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil {
		return nil, 0, product_errors.ErrProductInternal.WithInternal(err)
	}
	products := make([]*models.Product, 0, len(resp.Data))
	for _, p := range resp.Data {
		products = append(products, toModel(p))
	}
	total := len(products)
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return products, total, nil
}

// parseTime accepts the loose RFC3339-ish timestamp strings the product service
// emits and returns nil for empty strings. Anything that fails to parse is treated
// as nil rather than an error so the caller keeps a usable *models.Product with
// just the textual fields populated.
func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
