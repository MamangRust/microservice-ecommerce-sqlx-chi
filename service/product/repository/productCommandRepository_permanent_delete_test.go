package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
)

// productFKDriver simulates a Postgres connection whose Exec always fails
// with a foreign-key violation, so DeletePermanent / DeleteAll can be
// exercised on the failure path without a live database.
type productFKDriver struct{}

type productFKConn struct{}

func (productFKDriver) Open(name string) (driver.Conn, error) { return productFKConn{}, nil }

func (productFKConn) Prepare(query string) (driver.Stmt, error) { return productFKStmt{}, nil }
func (productFKConn) Close() error                              { return nil }
func (productFKConn) Begin() (driver.Tx, error)                 { return nil, errors.New("not implemented") }

type productFKStmt struct{}

func (productFKStmt) Close() error  { return nil }
func (productFKStmt) NumInput() int { return -1 }

func (productFKStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, &pgconn.PgError{Code: "23503", ConstraintName: "reviews_product_id_fkey"}
}

func (productFKStmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, errors.New("not implemented")
}

func init() {
	sql.Register("productFKStub", productFKDriver{})
}

func newProductCommandRepositoryStub(t *testing.T) *productCommandRepository {
	t.Helper()

	conn, err := sqlx.Connect("productFKStub", "stub://product-fk")
	if err != nil {
		t.Fatalf("sqlx.Connect: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return NewProductCommandRepository(conn)
}

func TestProductCommandRepositoryDeletePermanentReturnsConflictOnForeignKey(t *testing.T) {
	repo := newProductCommandRepositoryStub(t)

	deleted, err := repo.DeletePermanent(context.Background(), 7)
	assertProductPermanentDeleteConflict(t, deleted, err)
}

func TestProductCommandRepositoryDeleteAllReturnsConflictOnForeignKey(t *testing.T) {
	repo := newProductCommandRepositoryStub(t)

	deleted, err := repo.DeleteAll(context.Background())
	assertProductPermanentDeleteConflict(t, deleted, err)
}

func assertProductPermanentDeleteConflict(t *testing.T, deleted bool, err error) {
	t.Helper()

	if deleted {
		t.Fatal("delete reported success after a foreign-key violation")
	}

	var appErr *shared_errors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T %v, want AppError", err, err)
	}
	if appErr.Type != shared_errors.ErrorTypeConflict || appErr.Code != 409 {
		t.Fatalf("error type/code = %s/%d, want CONFLICT/409", appErr.Type, appErr.Code)
	}
}
