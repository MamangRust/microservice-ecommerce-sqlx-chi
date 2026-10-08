// Package stats_reader_test exercises the ecommerce stats-reader gRPC surface
// end to end: it seeds a real ClickHouse with known aggregates, starts the real
// gRPC server with all seven stats services registered, and asserts the values
// that come back over the wire.
package stats_reader_test

import (
	"testing"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	categorypb "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	orderpb "github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	transactionpb "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	statsreaderhandler "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/handler"
	categoryrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/category"
	orderrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/order"
	transactionrepo "github.com/MamangRust/microservice-ecommerce-grpc-stats-reader/repository/transaction"
	pkgclickhouse "github.com/MamangRust/microservice-ecommerce-pkg/clickhouse"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	statsYear  = 2026
	statsMonth = 1
)

type StatsReaderSuite struct {
	tests.BaseTestSuite

	chConn chDriver.Conn

	order         orderpb.OrderStatsServiceClient
	orderMerchant orderpb.OrderStatsByMerchantServiceClient

	category         categorypb.CategoryStatsServiceClient
	categoryByID     categorypb.CategoryStatsByIdServiceClient
	categoryMerchant categorypb.CategoryStatsByMerchantServiceClient

	tx         transactionpb.TransactionStatsServiceClient
	txMerchant transactionpb.TransactionStatsByMerchantServiceClient
}

func (s *StatsReaderSuite) SetupSuite() {
	// The reader only needs ClickHouse; the base harness stands up the shared
	// Postgres/Redis containers, which this suite does not use.
	s.BaseTestSuite.SetupSuite()

	conn, err := pkgclickhouse.NewClient(s.Log)
	s.Require().NoError(err)
	s.Require().NoError(pkgclickhouse.ApplySchema(s.Ctx, conn, s.Log))
	s.chConn = conn

	orderHandler := statsreaderhandler.NewOrderStatsHandler(orderrepo.NewRepository(conn), s.Log)
	categoryHandler := statsreaderhandler.NewCategoryStatsHandler(categoryrepo.NewRepository(conn), s.Log)
	transactionHandler := statsreaderhandler.NewTransactionStatsHandler(transactionrepo.NewRepository(conn), s.Log)

	server := grpc.NewServer()
	orderpb.RegisterOrderStatsServiceServer(server, orderHandler)
	orderpb.RegisterOrderStatsByMerchantServiceServer(server, orderHandler)
	categorypb.RegisterCategoryStatsServiceServer(server, categoryHandler)
	categorypb.RegisterCategoryStatsByIdServiceServer(server, categoryHandler)
	categorypb.RegisterCategoryStatsByMerchantServiceServer(server, categoryHandler)
	transactionpb.RegisterTransactionStatsServiceServer(server, transactionHandler)
	transactionpb.RegisterTransactionStatsByMerchantServiceServer(server, transactionHandler)

	client := s.GetConnection(s.RegisterServer(server))

	s.order = orderpb.NewOrderStatsServiceClient(client)
	s.orderMerchant = orderpb.NewOrderStatsByMerchantServiceClient(client)
	s.category = categorypb.NewCategoryStatsServiceClient(client)
	s.categoryByID = categorypb.NewCategoryStatsByIdServiceClient(client)
	s.categoryMerchant = categorypb.NewCategoryStatsByMerchantServiceClient(client)
	s.tx = transactionpb.NewTransactionStatsServiceClient(client)
	s.txMerchant = transactionpb.NewTransactionStatsByMerchantServiceClient(client)

	s.seedClickHouse()
}

func (s *StatsReaderSuite) TearDownSuite() {
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.BaseTestSuite.TearDownSuite()
}

// seedClickHouse writes a small fixture with known aggregates. Every row has a
// distinct event_id, so each row is its own ReplacingMergeTree key and the
// reader (which does not query with FINAL) sees exactly what was seeded:
//
//	order_events      order 1 (merchant 1, 1500) + order 2 (merchant 2, 2000)
//	order_item_events order 1 → category 5 (3 items / 300)
//	                  order 2 → category 6 (2 items / 200)
//	transaction_events 2 success (3000) + 1 failed (500)
func (s *StatsReaderSuite) seedClickHouse() {
	jan := func(day int) time.Time {
		return time.Date(statsYear, time.January, day, 12, 0, 0, 0, time.UTC)
	}
	id := func(key string) uuid.UUID {
		return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key))
	}

	orderInsert := `INSERT INTO order_events
		(event_id, order_id, user_id, merchant_id, total_price, created_at, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	s.exec(orderInsert, id("order-1"), uint64(1), uint64(1), uint64(1), int64(1500), jan(10), uint64(1))
	s.exec(orderInsert, id("order-2"), uint64(2), uint64(2), uint64(2), int64(2000), jan(11), uint64(1))

	itemInsert := `INSERT INTO order_item_events
		(event_id, order_item_id, order_id, merchant_id, product_id, category_id,
		 category_name, quantity, price, created_at, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	s.exec(itemInsert, id("item-1"), uint64(1), uint64(1), uint64(1), uint64(100), uint64(5), "Cat A", uint32(3), int64(100), jan(10), uint64(1))
	s.exec(itemInsert, id("item-2"), uint64(2), uint64(2), uint64(2), uint64(101), uint64(6), "Cat B", uint32(2), int64(100), jan(11), uint64(1))

	txInsert := `INSERT INTO transaction_events
		(event_id, transaction_id, order_id, merchant_id, payment_method, amount, status, created_at, event_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	s.exec(txInsert, id("tx-1"), uint64(1), uint64(1), uint64(1), "cash", int64(1000), "success", jan(10), uint64(1))
	s.exec(txInsert, id("tx-2"), uint64(2), uint64(2), uint64(2), "cash", int64(2000), "success", jan(11), uint64(1))
	s.exec(txInsert, id("tx-3"), uint64(3), uint64(2), uint64(2), "transfer", int64(500), "failed", jan(11), uint64(1))
}

func (s *StatsReaderSuite) exec(query string, args ...any) {
	s.Require().NoError(s.chConn.Exec(s.Ctx, query, args...))
}

func (s *StatsReaderSuite) TestOrderStats() {
	monthly, err := s.order.FindMonthlyTotalRevenue(s.Ctx, &orderpb.FindYearMonthTotalRevenue{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal("Jan", monthly.Data[0].Month)
	s.Equal(int32(3500), monthly.Data[0].TotalRevenue)

	yearly, err := s.order.FindYearlyTotalRevenue(s.Ctx, &orderpb.FindYearTotalRevenue{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 1)
	s.Equal(int32(3500), yearly.Data[0].TotalRevenue)

	monthlyOrders, err := s.order.FindMonthlyRevenue(s.Ctx, &orderpb.FindYearOrder{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(monthlyOrders.Data, 1)
	s.Equal(int32(2), monthlyOrders.Data[0].OrderCount)
	s.Equal(int32(3500), monthlyOrders.Data[0].TotalRevenue)
	s.Equal(int32(5), monthlyOrders.Data[0].TotalItemsSold)

	yearlyOrders, err := s.order.FindYearlyRevenue(s.Ctx, &orderpb.FindYearOrder{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlyOrders.Data, 1)
	s.Equal(int32(2), yearlyOrders.Data[0].OrderCount)
	s.Equal(int32(3500), yearlyOrders.Data[0].TotalRevenue)
	s.Equal(int32(5), yearlyOrders.Data[0].TotalItemsSold)
	s.Equal(int32(2), yearlyOrders.Data[0].UniqueProductsSold)
}

func (s *StatsReaderSuite) TestOrderStatsByMerchant() {
	monthly, err := s.orderMerchant.FindMonthlyTotalRevenueByMerchant(s.Ctx, &orderpb.FindYearMonthTotalRevenueByMerchant{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal(int32(1500), monthly.Data[0].TotalRevenue)

	yearly, err := s.orderMerchant.FindYearlyTotalRevenueByMerchant(s.Ctx, &orderpb.FindYearTotalRevenueByMerchant{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 1)
	s.Equal(int32(1500), yearly.Data[0].TotalRevenue)

	monthlyOrders, err := s.orderMerchant.FindMonthlyRevenueByMerchant(s.Ctx, &orderpb.FindYearOrderByMerchant{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(monthlyOrders.Data, 1)
	s.Equal(int32(1), monthlyOrders.Data[0].OrderCount)
	s.Equal(int32(1500), monthlyOrders.Data[0].TotalRevenue)
	s.Equal(int32(3), monthlyOrders.Data[0].TotalItemsSold)

	yearlyOrders, err := s.orderMerchant.FindYearlyRevenueByMerchant(s.Ctx, &orderpb.FindYearOrderByMerchant{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(yearlyOrders.Data, 1)
	s.Equal(int32(1), yearlyOrders.Data[0].OrderCount)
	s.Equal(int32(1500), yearlyOrders.Data[0].TotalRevenue)
}

func (s *StatsReaderSuite) TestCategoryStats() {
	monthly, err := s.category.FindMonthPrice(s.Ctx, &categorypb.FindYearCategory{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 2)

	byCategory := map[int32]*categorypb.CategoryMonthPriceResponse{}
	for _, row := range monthly.Data {
		byCategory[row.CategoryId] = row
	}
	s.Equal("Cat A", byCategory[5].CategoryName)
	s.Equal(int32(1), byCategory[5].OrderCount)
	s.Equal(int32(3), byCategory[5].ItemsSold)
	s.Equal(int32(300), byCategory[5].TotalRevenue)
	s.Equal("Cat B", byCategory[6].CategoryName)
	s.Equal(int32(1), byCategory[6].OrderCount)
	s.Equal(int32(2), byCategory[6].ItemsSold)
	s.Equal(int32(200), byCategory[6].TotalRevenue)

	yearly, err := s.category.FindYearPrice(s.Ctx, &categorypb.FindYearCategory{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 2)
	for _, row := range yearly.Data {
		if row.CategoryId == 5 {
			s.Equal(int32(300), row.TotalRevenue)
			s.Equal(int32(1), row.UniqueProductsSold)
		}
	}

	monthlyTotals, err := s.category.FindMonthlyTotalPrices(s.Ctx, &categorypb.FindYearMonthTotalPrices{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(monthlyTotals.Data, 1)
	s.Equal(int32(500), monthlyTotals.Data[0].TotalRevenue)

	yearlyTotals, err := s.category.FindYearlyTotalPrices(s.Ctx, &categorypb.FindYearTotalPrices{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlyTotals.Data, 1)
	s.Equal(int32(500), yearlyTotals.Data[0].TotalRevenue)
}

func (s *StatsReaderSuite) TestCategoryStatsById() {
	monthly, err := s.categoryByID.FindMonthPriceById(s.Ctx, &categorypb.FindYearCategoryById{
		CategoryId: 5, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal(int32(1), monthly.Data[0].OrderCount)
	s.Equal(int32(3), monthly.Data[0].ItemsSold)
	s.Equal(int32(300), monthly.Data[0].TotalRevenue)

	yearly, err := s.categoryByID.FindYearPriceById(s.Ctx, &categorypb.FindYearCategoryById{
		CategoryId: 5, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 1)
	s.Equal(int32(3), yearly.Data[0].ItemsSold)
	s.Equal(int32(300), yearly.Data[0].TotalRevenue)
	s.Equal(int32(1), yearly.Data[0].UniqueProductsSold)

	monthlyTotals, err := s.categoryByID.FindMonthlyTotalPricesById(s.Ctx, &categorypb.FindYearMonthTotalPriceById{
		Year: statsYear, Month: statsMonth, CategoryId: 5,
	})
	s.Require().NoError(err)
	s.Require().Len(monthlyTotals.Data, 1)
	s.Equal(int32(300), monthlyTotals.Data[0].TotalRevenue)

	yearlyTotals, err := s.categoryByID.FindYearlyTotalPricesById(s.Ctx, &categorypb.FindYearTotalPriceById{
		Year: statsYear, CategoryId: 5,
	})
	s.Require().NoError(err)
	s.Require().Len(yearlyTotals.Data, 1)
	s.Equal(int32(300), yearlyTotals.Data[0].TotalRevenue)
}

func (s *StatsReaderSuite) TestCategoryStatsByMerchant() {
	monthly, err := s.categoryMerchant.FindMonthPriceByMerchant(s.Ctx, &categorypb.FindYearCategoryByMerchant{
		MerchantId: 1, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(monthly.Data, 1)
	s.Equal(int32(5), monthly.Data[0].CategoryId)
	s.Equal(int32(3), monthly.Data[0].ItemsSold)
	s.Equal(int32(300), monthly.Data[0].TotalRevenue)

	yearly, err := s.categoryMerchant.FindYearPriceByMerchant(s.Ctx, &categorypb.FindYearCategoryByMerchant{
		MerchantId: 1, Year: statsYear,
	})
	s.Require().NoError(err)
	s.Require().Len(yearly.Data, 1)
	s.Equal(int32(300), yearly.Data[0].TotalRevenue)

	monthlyTotals, err := s.categoryMerchant.FindMonthlyTotalPricesByMerchant(s.Ctx, &categorypb.FindYearMonthTotalPriceByMerchant{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(monthlyTotals.Data, 1)
	s.Equal(int32(300), monthlyTotals.Data[0].TotalRevenue)

	yearlyTotals, err := s.categoryMerchant.FindYearlyTotalPricesByMerchant(s.Ctx, &categorypb.FindYearTotalPriceByMerchant{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(yearlyTotals.Data, 1)
	s.Equal(int32(300), yearlyTotals.Data[0].TotalRevenue)
}

func (s *StatsReaderSuite) TestTransactionStats() {
	success, err := s.tx.GetMonthlyAmountSuccess(s.Ctx, &transactionpb.MonthAmountTransactionRequest{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(success.Data, 1)
	s.Equal(int32(2), success.Data[0].TotalSuccess)
	s.Equal(int32(3000), success.Data[0].TotalAmount)

	yearlySuccess, err := s.tx.GetYearlyAmountSuccess(s.Ctx, &transactionpb.YearAmountTransactionRequest{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlySuccess.Data, 1)
	s.Equal(int32(2), yearlySuccess.Data[0].TotalSuccess)
	s.Equal(int32(3000), yearlySuccess.Data[0].TotalAmount)

	failed, err := s.tx.GetMonthlyAmountFailed(s.Ctx, &transactionpb.MonthAmountTransactionRequest{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(failed.Data, 1)
	s.Equal(int32(1), failed.Data[0].TotalFailed)
	s.Equal(int32(500), failed.Data[0].TotalAmount)

	yearlyFailed, err := s.tx.GetYearlyAmountFailed(s.Ctx, &transactionpb.YearAmountTransactionRequest{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlyFailed.Data, 1)
	s.Equal(int32(1), yearlyFailed.Data[0].TotalFailed)
	s.Equal(int32(500), yearlyFailed.Data[0].TotalAmount)

	methodSuccess, err := s.tx.GetMonthlyTransactionMethodSuccess(s.Ctx, &transactionpb.MonthMethodTransactionRequest{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(methodSuccess.Data, 1)
	s.Equal("cash", methodSuccess.Data[0].PaymentMethod)
	s.Equal(int32(2), methodSuccess.Data[0].TotalTransactions)
	s.Equal(int32(3000), methodSuccess.Data[0].TotalAmount)

	yearlyMethodSuccess, err := s.tx.GetYearlyTransactionMethodSuccess(s.Ctx, &transactionpb.YearMethodTransactionRequest{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlyMethodSuccess.Data, 1)
	s.Equal("cash", yearlyMethodSuccess.Data[0].PaymentMethod)
	s.Equal(int32(3000), yearlyMethodSuccess.Data[0].TotalAmount)

	methodFailed, err := s.tx.GetMonthlyTransactionMethodFailed(s.Ctx, &transactionpb.MonthMethodTransactionRequest{
		Year: statsYear, Month: statsMonth,
	})
	s.Require().NoError(err)
	s.Require().Len(methodFailed.Data, 1)
	s.Equal("transfer", methodFailed.Data[0].PaymentMethod)
	s.Equal(int32(1), methodFailed.Data[0].TotalTransactions)
	s.Equal(int32(500), methodFailed.Data[0].TotalAmount)

	yearlyMethodFailed, err := s.tx.GetYearlyTransactionMethodFailed(s.Ctx, &transactionpb.YearMethodTransactionRequest{Year: statsYear})
	s.Require().NoError(err)
	s.Require().Len(yearlyMethodFailed.Data, 1)
	s.Equal("transfer", yearlyMethodFailed.Data[0].PaymentMethod)
	s.Equal(int32(1), yearlyMethodFailed.Data[0].TotalTransactions)
	s.Equal(int32(500), yearlyMethodFailed.Data[0].TotalAmount)
}

func (s *StatsReaderSuite) TestTransactionStatsByMerchant() {
	success, err := s.txMerchant.GetMonthlyAmountSuccessByMerchant(s.Ctx, &transactionpb.MonthAmountTransactionMerchantRequest{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(success.Data, 1)
	s.Equal(int32(1), success.Data[0].TotalSuccess)
	s.Equal(int32(1000), success.Data[0].TotalAmount)

	yearlySuccess, err := s.txMerchant.GetYearlyAmountSuccessByMerchant(s.Ctx, &transactionpb.YearAmountTransactionMerchantRequest{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(yearlySuccess.Data, 1)
	s.Equal(int32(1), yearlySuccess.Data[0].TotalSuccess)
	s.Equal(int32(1000), yearlySuccess.Data[0].TotalAmount)

	// Merchant 1 has no failed transaction.
	failed, err := s.txMerchant.GetMonthlyAmountFailedByMerchant(s.Ctx, &transactionpb.MonthAmountTransactionMerchantRequest{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Empty(failed.Data)

	yearlyFailed, err := s.txMerchant.GetYearlyAmountFailedByMerchant(s.Ctx, &transactionpb.YearAmountTransactionMerchantRequest{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Empty(yearlyFailed.Data)

	methodSuccess, err := s.txMerchant.GetMonthlyTransactionMethodByMerchantSuccess(s.Ctx, &transactionpb.MonthMethodTransactionMerchantRequest{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(methodSuccess.Data, 1)
	s.Equal("cash", methodSuccess.Data[0].PaymentMethod)
	s.Equal(int32(1), methodSuccess.Data[0].TotalTransactions)
	s.Equal(int32(1000), methodSuccess.Data[0].TotalAmount)

	yearlyMethodSuccess, err := s.txMerchant.GetYearlyTransactionMethodByMerchantSuccess(s.Ctx, &transactionpb.YearMethodTransactionMerchantRequest{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Require().Len(yearlyMethodSuccess.Data, 1)
	s.Equal("cash", yearlyMethodSuccess.Data[0].PaymentMethod)
	s.Equal(int32(1000), yearlyMethodSuccess.Data[0].TotalAmount)

	methodFailed, err := s.txMerchant.GetMonthlyTransactionMethodByMerchantFailed(s.Ctx, &transactionpb.MonthMethodTransactionMerchantRequest{
		Year: statsYear, Month: statsMonth, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Empty(methodFailed.Data)

	yearlyMethodFailed, err := s.txMerchant.GetYearlyTransactionMethodByMerchantFailed(s.Ctx, &transactionpb.YearMethodTransactionMerchantRequest{
		Year: statsYear, MerchantId: 1,
	})
	s.Require().NoError(err)
	s.Empty(yearlyMethodFailed.Data)
}

// TestAllRegisteredServicesReachable proves every one of the seven services is
// wired to a handler: a missing registration would answer Unimplemented.
func (s *StatsReaderSuite) TestAllRegisteredServicesReachable() {
	cases := []struct {
		service string
		call    func() error
	}{
		{"OrderStatsService", func() error {
			_, err := s.order.FindYearlyTotalRevenue(s.Ctx, &orderpb.FindYearTotalRevenue{Year: statsYear})
			return err
		}},
		{"OrderStatsByMerchantService", func() error {
			_, err := s.orderMerchant.FindYearlyTotalRevenueByMerchant(s.Ctx, &orderpb.FindYearTotalRevenueByMerchant{Year: statsYear, MerchantId: 1})
			return err
		}},
		{"CategoryStatsService", func() error {
			_, err := s.category.FindYearlyTotalPrices(s.Ctx, &categorypb.FindYearTotalPrices{Year: statsYear})
			return err
		}},
		{"CategoryStatsByIdService", func() error {
			_, err := s.categoryByID.FindYearlyTotalPricesById(s.Ctx, &categorypb.FindYearTotalPriceById{Year: statsYear, CategoryId: 5})
			return err
		}},
		{"CategoryStatsByMerchantService", func() error {
			_, err := s.categoryMerchant.FindYearlyTotalPricesByMerchant(s.Ctx, &categorypb.FindYearTotalPriceByMerchant{Year: statsYear, MerchantId: 1})
			return err
		}},
		{"TransactionStatsService", func() error {
			_, err := s.tx.GetYearlyAmountSuccess(s.Ctx, &transactionpb.YearAmountTransactionRequest{Year: statsYear})
			return err
		}},
		{"TransactionStatsByMerchantService", func() error {
			_, err := s.txMerchant.GetYearlyAmountSuccessByMerchant(s.Ctx, &transactionpb.YearAmountTransactionMerchantRequest{Year: statsYear, MerchantId: 1})
			return err
		}},
	}

	for _, tc := range cases {
		s.Run(tc.service, func() {
			err := tc.call()
			s.NoError(err)
			s.NotEqual(codes.Unimplemented, status.Code(err), "%s is not registered", tc.service)
		})
	}
}

func TestStatsReaderSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	suite.Run(t, new(StatsReaderSuite))
}
