// Package backfill implements the stats-writer `backfill` command: it reads
// historical OLTP rows from the owning services (orders, order_items, products,
// categories, transactions) through their gRPC adapters and materializes them
// into ClickHouse through the same batch repository used for live events.
package backfill

import (
	"context"
	"fmt"
	"time"

	categoryadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/category"
	orderadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/transaction"
	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/repository"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/events"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// backfillPageSize is the chunk the backfill walks each source in. It matches the
// online services' PageSize and keeps memory bounded for large tables.
const backfillPageSize = 1000

// Backfiller reads OLTP events and pushes them into ClickHouse. Each source is
// reached through its owning service's adapter, so the backfill shares the exact
// same read path (and the same cross-context ownership rules) as the live
// pipeline.
type Backfiller struct {
	log          logger.LoggerInterface
	repo         repository.Repository
	orders       orderadapter.BulkRepository
	orderItems   orderitemadapter.BulkRepository
	products     productadapter.BulkRepository
	categories   categoryadapter.BulkRepository
	transactions transactionadapter.BulkRepository
}

// New builds a Backfiller from the ClickHouse repository and one BulkRepository
// per stats source. The sources are read through the owning services rather than
// touching PostgreSQL directly.
func New(
	log logger.LoggerInterface,
	repo repository.Repository,
	orders orderadapter.BulkRepository,
	orderItems orderitemadapter.BulkRepository,
	products productadapter.BulkRepository,
	categories categoryadapter.BulkRepository,
	transactions transactionadapter.BulkRepository,
) *Backfiller {
	return &Backfiller{
		log:          log,
		repo:         repo,
		orders:       orders,
		orderItems:   orderItems,
		products:     products,
		categories:   categories,
		transactions: transactions,
	}
}

// backfillEventID derives a deterministic UUID per entity so re-running the
// backfill replaces the same ReplacingMergeTree key (with a newer version)
// instead of appending duplicates.
func backfillEventID(kind string, id int32) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(fmt.Sprintf("backfill:%s:%d", kind, id))).String()
}

func eventTimeOf(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// Run streams all stats sources into ClickHouse. The event version is the
// backfill run timestamp so re-running supersedes previous rows.
func (b *Backfiller) Run(ctx context.Context) error {
	version := uint64(time.Now().Unix())
	counts := map[string]int{}

	if err := b.backfillOrders(ctx, version, counts); err != nil {
		return err
	}
	if err := b.backfillOrderItems(ctx, version, counts); err != nil {
		return err
	}
	if err := b.backfillTransactions(ctx, version, counts); err != nil {
		return err
	}

	if err := b.repo.Flush(ctx); err != nil {
		return fmt.Errorf("flush backfill batches: %w", err)
	}

	b.log.Info("backfill complete",
		zap.Int("orders", counts["order"]),
		zap.Int("order_items", counts["order_item"]),
		zap.Int("transactions", counts["transaction"]),
	)
	return nil
}

func (b *Backfiller) backfillOrders(ctx context.Context, version uint64, counts map[string]int) error {
	for page := 1; ; page++ {
		orders, _, err := b.orders.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query orders: %w", err)
		}
		for _, o := range orders {
			event := events.OrderEvent{
				OrderID:    o.OrderID,
				UserID:     o.UserID,
				MerchantID: o.MerchantID,
				TotalPrice: o.TotalPrice,
				Status:     "created",
				EventTime:  eventTimeOf(o.CreatedAt),
			}
			if err := b.repo.InsertOrderEvent(ctx, backfillEventID("order", o.OrderID), version, event); err != nil {
				return fmt.Errorf("insert order %d: %w", o.OrderID, err)
			}
			counts["order"]++
		}
		if len(orders) < backfillPageSize {
			break
		}
	}
	return nil
}

// backfillOrderItems denormalizes the catalog onto each order item: it walks
// orders (to recover the owning merchant) and products/categories (to recover
// the category id and name), then materializes every order item.
func (b *Backfiller) backfillOrderItems(ctx context.Context, version uint64, counts map[string]int) error {
	orderMerchant := map[int32]int32{}
	for page := 1; ; page++ {
		orders, _, err := b.orders.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query orders for merchant map: %w", err)
		}
		for _, o := range orders {
			orderMerchant[o.OrderID] = o.MerchantID
		}
		if len(orders) < backfillPageSize {
			break
		}
	}

	productCategory := map[int32]int32{}
	for page := 1; ; page++ {
		products, _, err := b.products.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query products for category map: %w", err)
		}
		for _, p := range products {
			productCategory[p.ProductID] = p.CategoryID
		}
		if len(products) < backfillPageSize {
			break
		}
	}

	categoryName := map[int32]string{}
	for page := 1; ; page++ {
		categories, _, err := b.categories.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query categories for name map: %w", err)
		}
		for _, c := range categories {
			categoryName[c.CategoryID] = c.Name
		}
		if len(categories) < backfillPageSize {
			break
		}
	}

	for page := 1; ; page++ {
		items, _, err := b.orderItems.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query order items: %w", err)
		}
		for _, it := range items {
			catID := productCategory[it.ProductID]
			event := events.OrderItemEvent{
				OrderItemID:  it.OrderItemID,
				OrderID:      it.OrderID,
				MerchantID:   orderMerchant[it.OrderID],
				ProductID:    it.ProductID,
				CategoryID:   catID,
				CategoryName: categoryName[catID],
				Quantity:     it.Quantity,
				Price:        it.Price,
				EventTime:    eventTimeOf(it.CreatedAt),
			}
			if err := b.repo.InsertOrderItemEvent(ctx, backfillEventID("order_item", it.OrderItemID), version, event); err != nil {
				return fmt.Errorf("insert order item %d: %w", it.OrderItemID, err)
			}
			counts["order_item"]++
		}
		if len(items) < backfillPageSize {
			break
		}
	}
	return nil
}

func (b *Backfiller) backfillTransactions(ctx context.Context, version uint64, counts map[string]int) error {
	for page := 1; ; page++ {
		txs, _, err := b.transactions.FindAll(ctx, page, backfillPageSize)
		if err != nil {
			return fmt.Errorf("query transactions: %w", err)
		}
		for _, t := range txs {
			event := events.TransactionEvent{
				TransactionID: t.TransactionID,
				OrderID:       t.OrderID,
				MerchantID:    t.MerchantID,
				PaymentMethod: t.PaymentMethod,
				Amount:        t.Amount,
				Status:        t.Status,
				EventTime:     eventTimeOf(t.CreatedAt),
			}
			if err := b.repo.InsertTransactionEvent(ctx, backfillEventID("transaction", t.TransactionID), version, event); err != nil {
				return fmt.Errorf("insert transaction %d: %w", t.TransactionID, err)
			}
			counts["transaction"]++
		}
		if len(txs) < backfillPageSize {
			break
		}
	}
	return nil
}
