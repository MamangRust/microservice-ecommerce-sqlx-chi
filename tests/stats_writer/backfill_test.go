// Package stats_writer_test exercises the ecommerce stats-writer backfill end
// to end: it seeds the owning PostgreSQL tables (catalog + sales contexts),
// starts the real order/order_item/product/category/transaction gRPC services,
// runs the real backfiller — which reads every source through those services —
// against a real ClickHouse, and asserts the materialized rows.
package stats_writer_test

import (
	"testing"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	pborder "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	pborder_item "github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	pbcategory "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	pbtransaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	categoryadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/category"
	orderadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/order_item"
	productadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/transaction"
	pkgclickhouse "github.com/MamangRust/microservice-ecommerce-pkg/clickhouse"
	backfill "github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/backfill"
	statsrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/repository"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

const (
	categoryAName = "Cat A"
	categoryBName = "Cat B"

	productOnePrice = 1000
	productTwoPrice = 2000

	orderOneTotal = 3000
	orderTwoTotal = 5000

	txOneAmount = 3000
	txTwoAmount = 5000
)

type BackfillSuite struct {
	tests.BaseTestSuite

	chConn chDriver.Conn
	repo   statsrepo.Repository

	orderOne int32
	orderTwo int32
}

func (s *BackfillSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// The backfill reads every source through the owning gRPC services, so bring
	// them up (with their dependency chain: role→user→merchant, category,
	// shipping-address, then product/order-item/order/transaction).
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupCategoryService()
	s.SetupShippingAddressService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupTransactionService()
	s.SetupOrderService()

	conn, err := pkgclickhouse.NewClient(s.Log)
	s.Require().NoError(err)
	s.Require().NoError(pkgclickhouse.ApplySchema(s.Ctx, conn, s.Log))

	s.chConn = conn
	s.repo = statsrepo.NewClickhouseRepository(conn, s.Log)

	s.seedFixtures()
}

func (s *BackfillSuite) TearDownSuite() {
	if s.repo != nil {
		_ = s.repo.Close()
	}
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.BaseTestSuite.TearDownSuite()
}

// seedFixtures writes the fixture rows into the table that owns each entity:
// categories/products into the catalog context, and orders/order_items/
// transactions into the sales context. The test harness collapses every bounded
// context onto a single PostgreSQL container, so one pool is enough here.
//
// The backfill itself never reads PostgreSQL — it goes through the owning
// services — so this is purely a data setup step. slug_product/slug_category
// are left NULL so the unique constraints never collide.
func (s *BackfillSuite) seedFixtures() {
	ctx := s.Ctx
	db := s.SQLxDB()
	now := time.Now().UTC()

	// Catalog context: categories → products.
	var catA, catB int32
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ($1, 'a', $2, $2)
		RETURNING category_id
	`, categoryAName, now).Scan(&catA))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ($1, 'b', $2, $2)
		RETURNING category_id
	`, categoryBName, now).Scan(&catB))

	var productOne, productTwo int32
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO products (merchant_id, category_id, name, price, count_in_stock, created_at, updated_at)
		VALUES ($1, $2, 'P1', $3, 100, $4, $4)
		RETURNING product_id
	`, int32(1), catA, productOnePrice, now).Scan(&productOne))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO products (merchant_id, category_id, name, price, count_in_stock, created_at, updated_at)
		VALUES ($1, $2, 'P2', $3, 100, $4, $4)
		RETURNING product_id
	`, int32(2), catB, productTwoPrice, now).Scan(&productTwo))

	// Sales context: two orders, one per merchant (so merchant_id is recovered
	// per order by the backfill).
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, merchant_id, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING order_id
	`, int32(1), int32(1), orderOneTotal, now).Scan(&s.orderOne))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, merchant_id, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING order_id
	`, int32(2), int32(2), orderTwoTotal, now).Scan(&s.orderTwo))

	// Sales context: one order item per order, pointing at the two products so
	// both catalog categories are exercised.
	_, err := db.ExecContext(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
		VALUES ($1, $2, 2, $3, $4, $4),
		       ($5, $6, 1, $7, $4, $4)
	`, s.orderOne, productOne, productOnePrice, now, s.orderTwo, productTwo, productTwoPrice)
	s.Require().NoError(err)

	// Sales context: one successful transaction per order.
	_, err = db.ExecContext(ctx, `
		INSERT INTO transactions (order_id, merchant_id, payment_method, amount, payment_status, created_at, updated_at)
		VALUES ($1, $2, 'cash', $3, 'success', $4, $4),
		       ($5, $6, 'transfer', $7, 'success', $4, $4)
	`, s.orderOne, int32(1), txOneAmount, now, s.orderTwo, int32(2), txTwoAmount)
	s.Require().NoError(err)
}

func (s *BackfillSuite) TestBackfillMaterializesEverySource() {
	orderRepo := orderadapter.NewBulkAdapter(pborder.NewOrderQueryServiceClient(s.Conns["order"]))
	orderItemRepo := orderitemadapter.NewBulkAdapter(pborder_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]))
	productRepo := productadapter.NewAdapter(
		pbproduct.NewProductQueryServiceClient(s.Conns["product"]),
		pbproduct.NewProductCommandServiceClient(s.Conns["product"]),
	)
	categoryRepo := categoryadapter.NewBulkAdapter(pbcategory.NewCategoryQueryServiceClient(s.Conns["category"]))
	transactionRepo := transactionadapter.NewBulkAdapter(pbtransaction.NewTransactionQueryServiceClient(s.Conns["transaction"]))

	bf := backfill.New(s.Log, s.repo, orderRepo, orderItemRepo, productRepo, categoryRepo, transactionRepo)
	s.Require().NoError(bf.Run(s.Ctx))

	s.Equal(uint64(2), s.count(`SELECT count() FROM order_events`))
	s.Equal(int64(orderOneTotal+orderTwoTotal), s.sum(`SELECT sum(total_price) FROM order_events`))

	s.Equal(uint64(2), s.count(`SELECT count() FROM order_item_events`))
	s.Equal(uint64(0), s.count(
		`SELECT count() FROM order_item_events WHERE quantity <= 0 OR price <= 0`),
		"item quantity and price must be carried over")

	// Category must be denormalized from the catalog onto the order item.
	s.Equal(uint64(1), s.count(`SELECT count() FROM order_item_events WHERE category_name = 'Cat A'`))
	s.Equal(uint64(1), s.count(`SELECT count() FROM order_item_events WHERE category_name = 'Cat B'`))

	// Merchant must be recovered from the owning order.
	s.Equal(uint64(1), s.count(`SELECT count() FROM order_item_events WHERE merchant_id = 1`))
	s.Equal(uint64(1), s.count(`SELECT count() FROM order_item_events WHERE merchant_id = 2`))

	s.Equal(uint64(2), s.count(`SELECT count() FROM transaction_events`))
	s.Equal(uint64(2), s.count(`SELECT count() FROM transaction_events WHERE status = 'success'`))
	s.Equal(int64(txOneAmount+txTwoAmount), s.sum(`SELECT sum(amount) FROM transaction_events`))
	s.Equal(uint64(1), s.count(`SELECT count() FROM transaction_events WHERE payment_method = 'cash'`))
	s.Equal(uint64(1), s.count(`SELECT count() FROM transaction_events WHERE payment_method = 'transfer'`))
}

func (s *BackfillSuite) count(query string) uint64 {
	var n uint64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx, query).Scan(&n))
	return n
}

func (s *BackfillSuite) sum(query string) int64 {
	var n int64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx, query).Scan(&n))
	return n
}

func TestBackfillSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	suite.Run(t, new(BackfillSuite))
}
