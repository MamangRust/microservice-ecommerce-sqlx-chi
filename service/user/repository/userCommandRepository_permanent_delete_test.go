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

// userFailingConnector is a driver.Connector whose connections always fail with
// a foreign-key violation, letting the 23503 error mapping be tested without a
// live database.
type userFailingConnector struct{}

func (c *userFailingConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, &pgconn.PgError{Code: "23503", ConstraintName: "merchants_user_id_fkey"}
}

func (c *userFailingConnector) Driver() driver.Driver { return userFailingDriver{} }

type userFailingDriver struct{}

func (userFailingDriver) Open(string) (driver.Conn, error) {
	return nil, &pgconn.PgError{Code: "23503", ConstraintName: "merchants_user_id_fkey"}
}

func newUserFailingDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db := sqlx.NewDb(sql.OpenDB(&userFailingConnector{}), "pgx")
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserCommandRepositoryDeletePermanentReturnsConflictOnForeignKey(t *testing.T) {
	repo := NewUserCommandRepository(newUserFailingDB(t))

	deleted, err := repo.DeletePermanent(context.Background(), 7)
	assertUserPermanentDeleteConflict(t, deleted, err)
}

func TestUserCommandRepositoryDeleteAllReturnsConflictOnForeignKey(t *testing.T) {
	repo := NewUserCommandRepository(newUserFailingDB(t))

	deleted, err := repo.DeleteAll(context.Background())
	assertUserPermanentDeleteConflict(t, deleted, err)
}

func assertUserPermanentDeleteConflict(t *testing.T, deleted bool, err error) {
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
