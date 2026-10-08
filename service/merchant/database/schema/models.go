// Hand-written package db — migrated from sqlc-generated sources.
// Struct fields are unchanged from the generated models; db tags were added
// for sqlx scanning while json tags are preserved.

package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Merchant struct {
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	UserID       int32            `db:"user_id" json:"user_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Address      *string          `db:"address" json:"address"`
	ContactEmail *string          `db:"contact_email" json:"contact_email"`
	ContactPhone *string          `db:"contact_phone" json:"contact_phone"`
	Status       string           `db:"status" json:"status"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type MerchantDocument struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type OutboxEvent struct {
	OutboxID      int64            `db:"outbox_id" json:"outbox_id"`
	Topic         string           `db:"topic" json:"topic"`
	EventKey      string           `db:"event_key" json:"event_key"`
	Payload       []byte           `db:"payload" json:"payload"`
	Status        string           `db:"status" json:"status"`
	Attempts      int32            `db:"attempts" json:"attempts"`
	NextAttemptAt time.Time        `db:"next_attempt_at" json:"next_attempt_at"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
