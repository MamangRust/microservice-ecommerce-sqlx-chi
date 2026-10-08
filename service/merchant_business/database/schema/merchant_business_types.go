package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateMerchantBusinessInformationRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantBusinessInformationRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantsBusinessInformationRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	MerchantName           string           `json:"merchant_name" db:"merchant_name"`
	TotalCount             int64            `json:"total_count" db:"total_count"`
}

type GetMerchantsBusinessInformationActiveRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt              pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName           string           `json:"merchant_name" db:"merchant_name"`
	TotalCount             int64            `json:"total_count" db:"total_count"`
}

type GetMerchantsBusinessInformationTrashedRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt              pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName           string           `json:"merchant_name" db:"merchant_name"`
	TotalCount             int64            `json:"total_count" db:"total_count"`
}

type UpdateMerchantBusinessInformationRow struct {
	MerchantBusinessInfoID int32            `json:"merchant_business_info_id" db:"merchant_business_info_id"`
	MerchantID             int32            `json:"merchant_id" db:"merchant_id"`
	BusinessType           *string          `json:"business_type" db:"business_type"`
	TaxID                  *string          `json:"tax_id" db:"tax_id"`
	EstablishedYear        *int32           `json:"established_year" db:"established_year"`
	NumberOfEmployees      *int32           `json:"number_of_employees" db:"number_of_employees"`
	WebsiteUrl             *string          `json:"website_url" db:"website_url"`
	CreatedAt              pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt              pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}
