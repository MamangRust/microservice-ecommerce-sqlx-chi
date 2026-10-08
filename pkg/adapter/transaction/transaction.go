package transaction

import (
	"context"
	"fmt"
	"time"

	pbtransaction "github.com/MamangRust/microservice-ecommerce-grpc-pb/transaction"
	"github.com/MamangRust/microservice-ecommerce-pkg/adapter"
	"github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	transaction_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/transaction_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CommandRepository is the contract consumers depend on for transaction writes.
type CommandRepository interface {
	DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error)
	DeleteAll(ctx context.Context) (bool, error)
}

// QueryRepository is the read contract for transactions.
type QueryRepository interface {
	FindByID(ctx context.Context, transactionID int) (*models.Transaction, error)
}

// BulkRepository enumerates transactions page by page for the stats backfill.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error)
}

type commandAdapter struct {
	client pbtransaction.TransactionCommandServiceClient
	guard  *resilience.DependencyGuard
}

type queryAdapter struct {
	client pbtransaction.TransactionQueryServiceClient
	guard  *resilience.DependencyGuard
}

// NewCommandAdapter wraps the generated transaction command client. Zero or more
// guard options attach a per-call timeout, bulkhead and circuit breaker around
// every outbound gRPC call; with no options the guard stays nil and calls pass
// straight through.
func NewCommandAdapter(client pbtransaction.TransactionCommandServiceClient, opts ...adapter.GuardOption) CommandRepository {
	a := &commandAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// NewQueryAdapter wraps the generated transaction query client.
func NewQueryAdapter(client pbtransaction.TransactionQueryServiceClient, opts ...adapter.GuardOption) QueryRepository {
	a := &queryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// NewBulkAdapter wraps the generated transaction query client for bulk enumeration.
func NewBulkAdapter(client pbtransaction.TransactionQueryServiceClient, opts ...adapter.GuardOption) BulkRepository {
	a := &queryAdapter{client: client}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *commandAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *queryAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func (a *commandAdapter) DeleteByOrderIDPermanent(ctx context.Context, orderID int) (bool, error) {
	if a.client == nil {
		return false, fmt.Errorf("transaction command client is not initialized")
	}
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		_, callErr := a.client.DeleteTransactionByOrderPermanent(ctx, &pbtransaction.FindByIdTransactionRequest{
			Id: int32(orderID),
		})
		return callErr
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *commandAdapter) DeleteAll(ctx context.Context) (bool, error) {
	if a.client == nil {
		return false, fmt.Errorf("transaction command client is not initialized")
	}
	var res *pbtransaction.ApiResponseTransactionAll
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.DeleteAllTransactionPermanent(ctx, &emptypb.Empty{})
		return callErr
	})
	if err != nil {
		return false, err
	}

	return res.Status == "success", nil
}

func (a *queryAdapter) FindByID(ctx context.Context, transactionID int) (*models.Transaction, error) {
	var res *pbtransaction.ApiResponseTransaction
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		res, callErr = a.client.FindById(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(transactionID)})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return toModel(res.Data), nil
}

// FindAll walks every transaction page by page for the backfill. The returned int
// is the total record count reported by the service.
func (a *queryAdapter) FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error) {
	var resp *pbtransaction.ApiResponsePaginationTransaction
	err := a.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = a.client.FindAllTransactions(ctx, &pbtransaction.FindAllTransactionRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil {
		return nil, 0, transaction_errors.ErrFindAllTransactions.WithInternal(err)
	}
	transactions := make([]*models.Transaction, 0, len(resp.Data))
	for _, t := range resp.Data {
		transactions = append(transactions, toModel(t))
	}
	total := len(transactions)
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return transactions, total, nil
}

func toModel(t *pbtransaction.TransactionResponse) *models.Transaction {
	if t == nil {
		return nil
	}
	return &models.Transaction{
		TransactionID: t.Id,
		OrderID:       t.OrderId,
		MerchantID:    t.MerchantId,
		PaymentMethod: t.PaymentMethod,
		Amount:        t.Amount,
		Status:        t.PaymentStatus,
		CreatedAt:     parseTime(t.CreatedAt),
	}
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z"} {
		if parsed, err := time.Parse(layout, s); err == nil {
			return &parsed
		}
	}
	return nil
}
