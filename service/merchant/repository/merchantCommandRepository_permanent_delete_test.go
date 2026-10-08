package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	shared_errors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// fakeExecDB is a minimal database/sql driver that fails Exec with a canned
// error so repository error mapping can be tested without a real database.
type fakeExecDB struct {
	execErr error
	lastSQL string
}

func (f *fakeExecDB) Connect(ctx context.Context) (driver.Conn, error) { return f, nil }

func (f *fakeExecDB) Driver() driver.Driver { return f }

func (f *fakeExecDB) Open(name string) (driver.Conn, error) { return f, nil }

func (f *fakeExecDB) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("fakeExecDB: Prepare not supported")
}

func (f *fakeExecDB) Close() error { return nil }

func (f *fakeExecDB) Begin() (driver.Tx, error) {
	return nil, errors.New("fakeExecDB: Begin not supported")
}

func (f *fakeExecDB) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	f.lastSQL = query
	return nil, f.execErr
}

// newFakeExecSQLxDB wraps fakeExecDB in a real *sqlx.DB so the repository
// exercises the exact sqlx code path used in production.
func newFakeExecSQLxDB(t *testing.T, execErr error) (*sqlx.DB, *fakeExecDB) {
	t.Helper()
	f := &fakeExecDB{execErr: execErr}
	return sqlx.NewDb(sql.OpenDB(f), "postgres"), f
}

func TestMerchantCommandRepositoryDeletePermanentReturnsConflictOnForeignKey(t *testing.T) {
	conn, _ := newFakeExecSQLxDB(t, &pgconn.PgError{Code: "23503", ConstraintName: "products_merchant_id_fkey"})

	repo := NewMerchantCommandRepository(conn)

	deleted, err := repo.DeletePermanent(context.Background(), 7)
	assertMerchantPermanentDeleteConflict(t, deleted, err)
}

func TestMerchantCommandRepositoryDeleteAllReturnsConflictOnForeignKey(t *testing.T) {
	conn, _ := newFakeExecSQLxDB(t, &pgconn.PgError{Code: "23503", ConstraintName: "products_merchant_id_fkey"})

	repo := NewMerchantCommandRepository(conn)

	deleted, err := repo.DeleteAll(context.Background())
	assertMerchantPermanentDeleteConflict(t, deleted, err)
}

func assertMerchantPermanentDeleteConflict(t *testing.T, deleted bool, err error) {
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
