package transaction_test

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"net/http"
	"testing"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	tran_cache "github.com/MamangRust/microservice-ecommerce-grpc-transaction/cache"
	db "github.com/MamangRust/microservice-ecommerce-grpc-transaction/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/repository"
	"github.com/MamangRust/microservice-ecommerce-grpc-transaction/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	app_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ── Fault-injecting repository wrappers ─────────────────────────────────────
// Each wrapper delegates to the real repository and can inject a transport
// failure (gRPC Unavailable) or a repository error on demand, mirroring the
// payment-gateway failure-injection template.

type faultInjectingUserQuery struct {
	inner repository.UserQueryRepository
	fail  bool
}

func (f *faultInjectingUserQuery) FindByID(ctx context.Context, userID int) (*models.User, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "user service unavailable (injected)")
	}
	return f.inner.FindByID(ctx, userID)
}

func (f *faultInjectingUserQuery) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return f.inner.FindByEmail(ctx, email)
}

func (f *faultInjectingUserQuery) FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error) {
	return f.inner.FindByEmailAndVerify(ctx, email)
}

func (f *faultInjectingUserQuery) FindByVerificationCode(ctx context.Context, code string) (*models.User, error) {
	return f.inner.FindByVerificationCode(ctx, code)
}

type faultInjectingMerchantQuery struct {
	inner repository.MerchantQueryRepository
	fail  bool
}

func (f *faultInjectingMerchantQuery) FindByID(ctx context.Context, merchantID int) (*models.Merchant, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "merchant service unavailable (injected)")
	}
	return f.inner.FindByID(ctx, merchantID)
}

type faultInjectingOrderQuery struct {
	inner repository.OrderQueryRepository
	fail  bool
}

func (f *faultInjectingOrderQuery) FindByID(ctx context.Context, orderID int) (*models.Order, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "order service unavailable (injected)")
	}
	return f.inner.FindByID(ctx, orderID)
}

type faultInjectingOrderItemQuery struct {
	inner repository.OrderItemRepository
	fail  bool
}

func (f *faultInjectingOrderItemQuery) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "order-item service unavailable (injected)")
	}
	return f.inner.FindOrderItemByOrder(ctx, orderID)
}

func (f *faultInjectingOrderItemQuery) CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error) {
	return f.inner.CalculateTotalPrice(ctx, orderID)
}

type faultInjectingShippingQuery struct {
	inner repository.ShippingAddressQueryRepository
	fail  bool
}

func (f *faultInjectingShippingQuery) FindByID(ctx context.Context, shippingID int) (*models.ShippingAddress, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "shipping-address service unavailable (injected)")
	}
	return f.inner.FindByID(ctx, shippingID)
}

func (f *faultInjectingShippingQuery) FindByOrder(ctx context.Context, orderID int) (*models.ShippingAddress, error) {
	if f.fail {
		return nil, status.Error(codes.Unavailable, "shipping-address service unavailable (injected)")
	}
	return f.inner.FindByOrder(ctx, orderID)
}

// faultInjectingTransactionCommand can fail the durable insert (CreateInTx /
// Create) so the failure-injection suite can verify the error path of the
// transactional outbox: when the insert fails, no business row is committed.
type faultInjectingTransactionCommand struct {
	inner repository.TransactionCommandRepository
	fail  bool
}

func (f *faultInjectingTransactionCommand) Create(ctx context.Context, request *requests.CreateTransactionRequest) (*db.CreateTransactionRow, error) {
	if f.fail {
		return nil, app_errors.ErrInternal.WithMessage("transaction insert failed (injected)")
	}
	return f.inner.Create(ctx, request)
}

func (f *faultInjectingTransactionCommand) CreateInTx(ctx context.Context, tx *sqlx.Tx, request *requests.CreateTransactionRequest) (*db.CreateTransactionRow, error) {
	if f.fail {
		return nil, app_errors.ErrInternal.WithMessage("transaction insert failed (injected)")
	}
	return f.inner.CreateInTx(ctx, tx, request)
}

func (f *faultInjectingTransactionCommand) Update(ctx context.Context, request *requests.UpdateTransactionRequest) (*db.UpdateTransactionRow, error) {
	return f.inner.Update(ctx, request)
}

func (f *faultInjectingTransactionCommand) Trash(ctx context.Context, transactionID int) (*db.Transaction, error) {
	return f.inner.Trash(ctx, transactionID)
}

func (f *faultInjectingTransactionCommand) Restore(ctx context.Context, transactionID int) (*db.Transaction, error) {
	return f.inner.Restore(ctx, transactionID)
}

func (f *faultInjectingTransactionCommand) DeletePermanent(ctx context.Context, transactionID int) (bool, error) {
	return f.inner.DeletePermanent(ctx, transactionID)
}

func (f *faultInjectingTransactionCommand) DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error) {
	return f.inner.DeleteByOrderIDPermanent(ctx, orderID)
}

func (f *faultInjectingTransactionCommand) RestoreAll(ctx context.Context) (bool, error) {
	return f.inner.RestoreAll(ctx)
}

func (f *faultInjectingTransactionCommand) DeleteAll(ctx context.Context) (bool, error) {
	return f.inner.DeleteAll(ctx)
}

// faultInjectingOutbox can fail the outbox enqueue (CreateInTx) so the suite
// can verify that a failed event enqueue rolls back the business insert in the
// transactional path — no orphan transaction row, no lost event.
type faultInjectingOutbox struct {
	inner repository.OutboxRepository
	fail  bool
}

func (f *faultInjectingOutbox) Create(ctx context.Context, topic, key string, payload []byte) (*db.OutboxEvent, error) {
	if f.fail {
		return nil, app_errors.ErrInternal.WithMessage("outbox enqueue failed (injected)")
	}
	return f.inner.Create(ctx, topic, key, payload)
}

func (f *faultInjectingOutbox) CreateInTx(ctx context.Context, tx *sqlx.Tx, topic, key string, payload []byte) (*db.OutboxEvent, error) {
	if f.fail {
		return nil, app_errors.ErrInternal.WithMessage("outbox enqueue failed (injected)")
	}
	return f.inner.CreateInTx(ctx, tx, topic, key, payload)
}

func (f *faultInjectingOutbox) GetPending(ctx context.Context, limit int) ([]*db.OutboxEvent, error) {
	return f.inner.GetPending(ctx, limit)
}

func (f *faultInjectingOutbox) Claim(ctx context.Context, limit int, leaseUntil time.Time) ([]*db.OutboxEvent, error) {
	return f.inner.Claim(ctx, limit, leaseUntil)
}

func (f *faultInjectingOutbox) MarkDelivered(ctx context.Context, outboxID int64) (*db.OutboxEvent, error) {
	return f.inner.MarkDelivered(ctx, outboxID)
}

func (f *faultInjectingOutbox) MarkFailed(ctx context.Context, outboxID int64, nextAttemptAt time.Time) (*db.OutboxEvent, error) {
	return f.inner.MarkFailed(ctx, outboxID, nextAttemptAt)
}

func (f *faultInjectingOutbox) MarkDead(ctx context.Context, outboxID int64) (*db.OutboxEvent, error) {
	return f.inner.MarkDead(ctx, outboxID)
}

func (f *faultInjectingOutbox) DeleteOld(ctx context.Context, cutoff time.Time) (int64, error) {
	return f.inner.DeleteOld(ctx, cutoff)
}

type TransactionFailureInjectionTestSuite struct {
	tests.BaseTestSuite
	svc            *service.Service
	queries        *sqlx.DB
	userQuery      *faultInjectingUserQuery
	merchantQuery  *faultInjectingMerchantQuery
	orderQuery     *faultInjectingOrderQuery
	orderItemQuery *faultInjectingOrderItemQuery
	shippingQuery  *faultInjectingShippingQuery
	txCommand      *faultInjectingTransactionCommand
	outbox         *faultInjectingOutbox
	userID         int
	merchantID     int
	categoryID     int
	productID      int
	orderID        int
}

func (s *TransactionFailureInjectionTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupShippingAddressService()
	s.SetupOrderItemService()
	s.SetupOrderService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	s.queries = s.SQLxDB()

	// Real repositories, wrapped with fault injection per dependency.
	real := repository.NewRepositories(&repository.Deps{
		DB:                   s.queries,
		UserQueryClient:      pb_user.NewUserQueryServiceClient(s.Conns["user"]),
		MerchantQueryClient:  pb_merchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		OrderQueryClient:     pb_order.NewOrderQueryServiceClient(s.Conns["order"]),
		OrderItemQueryClient: pb_order_item.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		ShippingQueryClient:  pb_shipping_address.NewShippingQueryServiceClient(s.Conns["shipping-address"]),
		Guards: repository.GuardOptions{
			User:      []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log))},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log))},
			Order:     []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, s.Log))},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log))},
			Shipping:  []adapter.GuardOption{adapter.WithDependencyGuard(resilience.NewDependencyGuard("shipping_address", 5, 30, 100, 3*time.Second, s.Log))},
		},
	})

	s.userQuery = &faultInjectingUserQuery{inner: real.UserQuery}
	s.merchantQuery = &faultInjectingMerchantQuery{inner: real.MerchantQuery}
	s.orderQuery = &faultInjectingOrderQuery{inner: real.OrderQuery}
	s.orderItemQuery = &faultInjectingOrderItemQuery{inner: real.OrderItem}
	s.shippingQuery = &faultInjectingShippingQuery{inner: real.ShippingAddress}
	s.txCommand = &faultInjectingTransactionCommand{inner: real.TransactionCommand}
	s.outbox = &faultInjectingOutbox{inner: real.Outbox}

	mencache := tran_cache.NewMencache(cacheStore)
	s.svc = service.NewService(&service.Deps{
		Kafka: nil,
		Pool:  s.SQLxDB(), // transactional outbox path (single commit)
		Cache: mencache,
		Repositories: &repository.Repositories{
			TransactionCommand: s.txCommand,
			TransactionQuery:   real.TransactionQuery,
			OrderItem:          s.orderItemQuery,
			OrderQuery:         s.orderQuery,
			MerchantQuery:      s.merchantQuery,
			ShippingAddress:    s.shippingQuery,
			UserQuery:          s.userQuery,
			Outbox:             s.outbox,
		},
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Seed the full dependency chain once; failure tests reuse it.
	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	s.categoryID = s.SeedCategory(ctx)
	s.productID = s.SeedProduct(ctx, s.merchantID, s.categoryID)
	s.orderID = s.SeedOrder(ctx, s.userID, s.merchantID, s.productID)
	s.SeedOrderItem(ctx, s.orderID, s.productID)
	s.SeedShippingAddress(ctx, s.orderID)
}

func (s *TransactionFailureInjectionTestSuite) createReq(amount int) *requests.CreateTransactionRequest {
	return &requests.CreateTransactionRequest{
		UserID:        s.userID,
		MerchantID:    s.merchantID,
		OrderID:       s.orderID,
		Amount:        amount,
		PaymentMethod: "Transfer Bank",
	}
}

// TestF1_InvalidAmount_RejectedWith400 verifies that an amount below the
// calculated total surfaces as a 400 (BadRequest), never a 500.
func (s *TransactionFailureInjectionTestSuite) TestF1_InvalidAmount_RejectedWith400() {
	ctx := context.Background()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1))
	s.Require().Error(err, "amount below total must be rejected")

	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr), "insufficient balance must be an AppError, got %T", err)
	s.Equal(http.StatusBadRequest, appErr.Code, "insufficient balance must map to 400")
}

// TestF2_UserLookupDown_PropagatesUnavailable verifies that a failing user
// dependency surfaces its gRPC status (Unavailable) without a nil dereference
// or panic.
func (s *TransactionFailureInjectionTestSuite) TestF2_UserLookupDown_PropagatesUnavailable() {
	ctx := context.Background()
	s.userQuery.fail = true
	defer func() { s.userQuery.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.True(ok, "dependency failure must be a gRPC status error, got %T", err)
	s.Equal(codes.Unavailable, st.Code(), "user service down must map to Unavailable, got %v", st.Code())
}

// TestF3_MerchantLookupDown_PropagatesUnavailable mirrors F2 for the merchant
// dependency.
func (s *TransactionFailureInjectionTestSuite) TestF3_MerchantLookupDown_PropagatesUnavailable() {
	ctx := context.Background()
	s.merchantQuery.fail = true
	defer func() { s.merchantQuery.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.True(ok, "dependency failure must be a gRPC status error, got %T", err)
	s.Equal(codes.Unavailable, st.Code(), "merchant service down must map to Unavailable, got %v", st.Code())
}

// TestF4_OrderLookupDown_PropagatesUnavailable mirrors F2 for the order
// dependency.
func (s *TransactionFailureInjectionTestSuite) TestF4_OrderLookupDown_PropagatesUnavailable() {
	ctx := context.Background()
	s.orderQuery.fail = true
	defer func() { s.orderQuery.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.True(ok, "dependency failure must be a gRPC status error, got %T", err)
	s.Equal(codes.Unavailable, st.Code(), "order service down must map to Unavailable, got %v", st.Code())
}

// TestF5_OrderItemLookupDown_PropagatesUnavailable mirrors F2 for the order
// item dependency.
func (s *TransactionFailureInjectionTestSuite) TestF5_OrderItemLookupDown_PropagatesUnavailable() {
	ctx := context.Background()
	s.orderItemQuery.fail = true
	defer func() { s.orderItemQuery.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.True(ok, "dependency failure must be a gRPC status error, got %T", err)
	s.Equal(codes.Unavailable, st.Code(), "order-item service down must map to Unavailable, got %v", st.Code())
}

// TestF6_ShippingLookupDown_PropagatesUnavailable mirrors F2 for the shipping
// address dependency.
func (s *TransactionFailureInjectionTestSuite) TestF6_ShippingLookupDown_PropagatesUnavailable() {
	ctx := context.Background()
	s.shippingQuery.fail = true
	defer func() { s.shippingQuery.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	st, ok := status.FromError(err)
	s.True(ok, "dependency failure must be a gRPC status error, got %T", err)
	s.Equal(codes.Unavailable, st.Code(), "shipping service down must map to Unavailable, got %v", st.Code())
}

// TestF7_TransactionInsertFailure_RollsBack verifies that a failed transaction
// insert surfaces as an error and leaves no durable business row behind (the
// transactional outbox rolls the unit back).
func (s *TransactionFailureInjectionTestSuite) TestF7_TransactionInsertFailure_RollsBack() {
	ctx := context.Background()
	s.txCommand.fail = true
	defer func() { s.txCommand.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err)

	var appErr *app_errors.AppError
	s.True(errors.As(err, &appErr), "insert failure must be an AppError, got %T", err)
	s.Equal(http.StatusInternalServerError, appErr.Code, "insert failure must map to 500")
}

// TestF8_OutboxEnqueueFailure_RollsBack is the financial-safety guarantee: when
// the outbox event enqueue fails inside the transaction, the whole unit rolls
// back — no orphan transaction row and no silently-lost event. The caller sees
// an error so it can retry safely.
func (s *TransactionFailureInjectionTestSuite) TestF8_OutboxEnqueueFailure_RollsBack() {
	ctx := context.Background()
	s.outbox.fail = true
	defer func() { s.outbox.fail = false }()

	_, err := s.svc.TransactionCommand.Create(ctx, s.createReq(1000000))
	s.Require().Error(err, "outbox enqueue failure must surface to the caller")

	// No transaction row may exist for the order: the business insert committed
	// together with the outbox events, and a failed enqueue aborts all of it.
	_, qErr := s.svc.TransactionQuery.FindByOrderID(ctx, s.orderID)
	s.Error(qErr, "no transaction row may exist after a rolled-back create")
	var appErr *app_errors.AppError
	s.True(errors.As(qErr, &appErr), "lookup after rollback must be a typed error, got %T", qErr)
	s.Equal(app_errors.ErrorTypeNotFound, appErr.Type, "rolled-back transaction must be absent (not-found), got %v", qErr)
}

func TestTransactionFailureInjectionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionFailureInjectionTestSuite))
}
