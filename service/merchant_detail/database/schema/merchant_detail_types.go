package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateMerchantDetailRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantDetailRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	SocialMediaLinks []byte           `json:"social_media_links" db:"social_media_links"`
}

type GetMerchantDetailTrashedRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type GetMerchantDetailsRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
	SocialMediaLinks []byte           `json:"social_media_links" db:"social_media_links"`
}

type GetMerchantDetailsActiveRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
	SocialMediaLinks []byte           `json:"social_media_links" db:"social_media_links"`
}

type GetMerchantDetailsTrashedRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
	MerchantName     string           `json:"merchant_name" db:"merchant_name"`
	TotalCount       int64            `json:"total_count" db:"total_count"`
	SocialMediaLinks []byte           `json:"social_media_links" db:"social_media_links"`
}

type UpdateMerchantDetailRow struct {
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	MerchantID       int32            `json:"merchant_id" db:"merchant_id"`
	DisplayName      *string          `json:"display_name" db:"display_name"`
	CoverImageUrl    *string          `json:"cover_image_url" db:"cover_image_url"`
	LogoUrl          *string          `json:"logo_url" db:"logo_url"`
	ShortDescription *string          `json:"short_description" db:"short_description"`
	WebsiteUrl       *string          `json:"website_url" db:"website_url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type CreateMerchantSocialMediaLinkRow struct {
	MerchantSocialID int32            `json:"merchant_social_id" db:"merchant_social_id"`
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	Platform         string           `json:"platform" db:"platform"`
	Url              string           `json:"url" db:"url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}

type UpdateMerchantSocialMediaLinkRow struct {
	MerchantSocialID int32            `json:"merchant_social_id" db:"merchant_social_id"`
	MerchantDetailID int32            `json:"merchant_detail_id" db:"merchant_detail_id"`
	Platform         string           `json:"platform" db:"platform"`
	Url              string           `json:"url" db:"url"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
}
