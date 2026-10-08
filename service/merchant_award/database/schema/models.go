package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type MerchantCertificationsAndAward struct {
	MerchantCertificationID int32            `json:"merchant_certification_id" db:"merchant_certification_id"`
	MerchantID              int32            `json:"merchant_id" db:"merchant_id"`
	Title                   string           `json:"title" db:"title"`
	Description             *string          `json:"description" db:"description"`
	IssuedBy                *string          `json:"issued_by" db:"issued_by"`
	IssueDate               pgtype.Date      `json:"issue_date" db:"issue_date"`
	ExpiryDate              pgtype.Date      `json:"expiry_date" db:"expiry_date"`
	CertificateUrl          *string          `json:"certificate_url" db:"certificate_url"`
	CreatedAt               pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt               pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt               pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
