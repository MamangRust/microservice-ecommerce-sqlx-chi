package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
	"github.com/jmoiron/sqlx"
)

type merchantDocumentQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantDocumentQueryRepository(db *sqlx.DB) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{
		db: db,
	}
}

const getMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET $3`

const getActiveMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NULL
    AND status != 'deleted'
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET $3`

const getTrashedMerchantDocuments = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at,
    deleted_at,
    COUNT(*) OVER () AS total_count
FROM merchant_documents
WHERE
    deleted_at IS NOT NULL
    AND (
        $1::TEXT IS NULL
        OR document_type ILIKE '%' || $1 || '%'
        OR status ILIKE '%' || $1 || '%'
        OR note ILIKE '%' || $1 || '%'
    )
ORDER BY document_id
LIMIT $2
OFFSET $3`

const getMerchantDocument = `SELECT
    document_id,
    merchant_id,
    document_type,
    document_url,
    status,
    note,
    uploaded_at,
    created_at,
    updated_at
FROM merchant_documents
WHERE
    document_id = $1
    AND deleted_at IS NULL`

func (r *merchantDocumentQueryRepository) FindAll(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &docs, getMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	var totalCount int
	if len(docs) > 0 {
		totalCount = int(docs[0].TotalCount)
	}

	return docs, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetActiveMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetActiveMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &docs, getActiveMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	var totalCount int
	if len(docs) > 0 {
		totalCount = int(docs[0].TotalCount)
	}

	return docs, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetTrashedMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var docs []*db.GetTrashedMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &docs, getTrashedMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	var totalCount int
	if len(docs) > 0 {
		totalCount = int(docs[0].TotalCount)
	}

	return docs, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindByID(ctx context.Context, id int) (*db.GetMerchantDocumentRow, error) {
	var doc db.GetMerchantDocumentRow
	err := r.db.GetContext(ctx, &doc, getMerchantDocument, int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchant_errors.ErrMerchantNotFound.WithInternal(err)
		}
		return nil, merchant_errors.ErrMerchantInternal.WithInternal(err)
	}

	return &doc, nil
}
