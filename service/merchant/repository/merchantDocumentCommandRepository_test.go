package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/jmoiron/sqlx"
)

// fakeQueryDB is a minimal database/sql driver that records the executed SQL
// and its positional args and returns a canned row for queries. It allows
// testing the repository param building without a real database.
type fakeQueryDB struct {
	lastSQL  string
	lastArgs []driver.Value

	columns []string
	row     []driver.Value
	rowErr  error
}

func (f *fakeQueryDB) Connect(ctx context.Context) (driver.Conn, error) { return f, nil }

func (f *fakeQueryDB) Driver() driver.Driver { return f }

func (f *fakeQueryDB) Open(name string) (driver.Conn, error) { return f, nil }

func (f *fakeQueryDB) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("fakeQueryDB: Prepare not supported")
}

func (f *fakeQueryDB) Close() error { return nil }

func (f *fakeQueryDB) Begin() (driver.Tx, error) {
	return nil, errors.New("fakeQueryDB: Begin not supported")
}

func (f *fakeQueryDB) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	f.lastSQL = query
	f.lastArgs = namedValues(args)
	return driver.RowsAffected(0), nil
}

func (f *fakeQueryDB) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	f.lastSQL = query
	f.lastArgs = namedValues(args)
	if f.rowErr != nil {
		return nil, f.rowErr
	}
	return &fakeQueryRows{columns: f.columns, vals: f.row}, nil
}

// fakeQueryRows returns a single canned row.
type fakeQueryRows struct {
	columns []string
	vals    []driver.Value
	done    bool
}

func (r *fakeQueryRows) Columns() []string { return r.columns }

func (r *fakeQueryRows) Close() error { return nil }

func (r *fakeQueryRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	copy(dest, r.vals)
	return nil
}

func namedValues(args []driver.NamedValue) []driver.Value {
	vals := make([]driver.Value, len(args))
	for i, a := range args {
		vals[i] = a.Value
	}
	return vals
}

// newFakeSQLxDB wraps fakeQueryDB in a real *sqlx.DB so the repository
// exercises the exact sqlx code path used in production.
func newFakeSQLxDB(t *testing.T, row []driver.Value, rowErr error) (*sqlx.DB, *fakeQueryDB) {
	t.Helper()
	f := &fakeQueryDB{
		columns: []string{
			"document_id", "merchant_id", "document_type", "document_url",
			"status", "note", "uploaded_at", "created_at", "updated_at",
		},
		row:    row,
		rowErr: rowErr,
	}
	return sqlx.NewDb(sql.OpenDB(f), "postgres"), f
}

// merchantDocumentRowFixture mirrors the column order of the
// updateMerchantDocument RETURNING scan.
func merchantDocumentRowFixture(docID, merchantID int) []driver.Value {
	return []driver.Value{
		int64(docID),
		int64(merchantID),
		"updated_type",
		"https://example.com/updated.pdf",
		"verified",
		"Approved",
		nil,
		nil,
		nil,
	}
}

func TestMerchantDocumentCommandRepositoryUpdateTargetsDocumentID(t *testing.T) {
	const (
		documentID = 42
		merchantID = 7
	)

	conn, fake := newFakeSQLxDB(t, merchantDocumentRowFixture(documentID, merchantID), nil)
	repo := NewMerchantDocumentCommandRepository(conn)
	note := "replace before review"

	updated, err := repo.Update(context.Background(), &requests.UpdateMerchantDocumentRequest{
		DocumentID:   intPtr(documentID),
		MerchantID:   merchantID,
		DocumentType: "business_license",
		DocumentUrl:  "https://example.com/license.pdf",
		Status:       "pending",
		Note:         note,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.DocumentID != int32(documentID) {
		t.Fatalf("updated document id = %d, want %d", updated.DocumentID, documentID)
	}

	assertMerchantDocumentUpdateTarget(t, fake, documentID, merchantID)
}

func TestMerchantDocumentCommandRepositoryUpdateStatusTargetsDocumentID(t *testing.T) {
	const (
		documentID = 84
		merchantID = 13
	)

	conn, fake := newFakeSQLxDB(t, merchantDocumentRowFixture(documentID, merchantID), nil)
	repo := NewMerchantDocumentCommandRepository(conn)
	note := "verified"

	updated, err := repo.UpdateStatus(context.Background(), &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: intPtr(documentID),
		MerchantID: merchantID,
		Status:     "approved",
		Note:       note,
	})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if updated.DocumentID != int32(documentID) {
		t.Fatalf("updated document id = %d, want %d", updated.DocumentID, documentID)
	}

	assertMerchantDocumentUpdateTarget(t, fake, documentID, merchantID)
}

func assertMerchantDocumentUpdateTarget(t *testing.T, fake *fakeQueryDB, documentID, merchantID int) {
	t.Helper()

	if !strings.Contains(fake.lastSQL, "document_id = $1") {
		t.Fatalf("update query target = %q, want document_id", fake.lastSQL)
	}
	if len(fake.lastArgs) == 0 {
		t.Fatal("update query received no arguments")
	}
	gotDocumentID, ok := fake.lastArgs[0].(int64)
	if !ok {
		t.Fatalf("first update argument type = %T, want int64", fake.lastArgs[0])
	}
	if gotDocumentID != int64(documentID) {
		t.Fatalf("first update argument = %d, want DocumentID %d (MerchantID is %d)", gotDocumentID, documentID, merchantID)
	}
	if gotDocumentID == int64(merchantID) {
		t.Fatalf("first update argument incorrectly used MerchantID %d", merchantID)
	}
}

// TestMerchantDocumentCommandRepositoryScanContext verifies the sqlx scan
// path maps the RETURNING columns onto the db.UpdateMerchantDocumentRow db tags.
func TestMerchantDocumentCommandRepositoryScanContext(t *testing.T) {
	conn, _ := newFakeSQLxDB(t, merchantDocumentRowFixture(42, 7), nil)
	repo := NewMerchantDocumentCommandRepository(conn)

	res, err := repo.Update(context.Background(), &requests.UpdateMerchantDocumentRequest{
		DocumentID:   intPtr(42),
		MerchantID:   7,
		DocumentType: "business_license",
		DocumentUrl:  "https://example.com/license.pdf",
		Status:       "pending",
		Note:         "ok",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if res.DocumentID != int32(42) {
		t.Fatalf("scanned DocumentID = %d, want 42", res.DocumentID)
	}
	if res.MerchantID != int32(7) {
		t.Fatalf("scanned MerchantID = %d, want 7", res.MerchantID)
	}
	if res.DocumentType != "updated_type" {
		t.Fatalf("scanned DocumentType = %q, want updated_type", res.DocumentType)
	}
}

func intPtr(value int) *int {
	return &value
}
