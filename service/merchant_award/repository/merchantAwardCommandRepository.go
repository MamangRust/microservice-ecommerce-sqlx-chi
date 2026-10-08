package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_award/database/schema"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	merchantaward_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant_award"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"
)

const createMerchantCertificationOrAward = `INSERT INTO
    merchant_certifications_and_awards (
        merchant_id,
        title,
        description,
        issued_by,
        issue_date,
        expiry_date,
        certificate_url
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    merchant_certification_id,
    merchant_id,
    title,
    description,
    issued_by,
    issue_date,
    expiry_date,
    certificate_url,
    created_at,
    updated_at`

const updateMerchantCertificationOrAward = `UPDATE merchant_certifications_and_awards
SET
    title = $2,
    description = $3,
    issued_by = $4,
    issue_date = $5,
    expiry_date = $6,
    certificate_url = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE
    merchant_certification_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_certification_id,
    merchant_id,
    title,
    description,
    issued_by,
    issue_date,
    expiry_date,
    certificate_url,
    created_at,
    updated_at`

const trashMerchantCertificationOrAward = `UPDATE merchant_certifications_and_awards
SET
    deleted_at = CURRENT_TIMESTAMP
WHERE
    merchant_certification_id = $1
    AND deleted_at IS NULL
RETURNING
    merchant_certification_id,
    merchant_id,
    title,
    description,
    issued_by,
    issue_date,
    expiry_date,
    certificate_url,
    created_at,
    updated_at,
    deleted_at`

const restoreMerchantCertificationOrAward = `UPDATE merchant_certifications_and_awards
SET
    deleted_at = NULL
WHERE
    merchant_certification_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    merchant_certification_id,
    merchant_id,
    title,
    description,
    issued_by,
    issue_date,
    expiry_date,
    certificate_url,
    created_at,
    updated_at,
    deleted_at`

const deleteMerchantCertificationOrAwardPermanently = `DELETE FROM merchant_certifications_and_awards
WHERE
    merchant_certification_id = $1
    AND deleted_at IS NOT NULL`

const restoreAllMerchantCertificationsAndAwards = `UPDATE merchant_certifications_and_awards
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentMerchantCertificationsAndAwards = `DELETE FROM merchant_certifications_and_awards
WHERE
    deleted_at IS NOT NULL`

type merchantAwardCommandRepository struct {
	db *sqlx.DB
}

func NewMerchantAwardCommandRepository(db *sqlx.DB) *merchantAwardCommandRepository {
	return &merchantAwardCommandRepository{
		db: db,
	}
}

func (r *merchantAwardCommandRepository) Create(
	ctx context.Context,
	request *requests.CreateMerchantCertificationOrAwardRequest,
) (*db.CreateMerchantCertificationOrAwardRow, error) {

	var award db.CreateMerchantCertificationOrAwardRow

	if err := r.db.GetContext(ctx, &award, createMerchantCertificationOrAward,
		int32(request.MerchantID),
		request.Title,
		stringPtr(request.Description),
		stringPtr(request.IssuedBy),
		parseDateToPgDate(request.IssueDate),
		parseDateToPgDate(request.ExpiryDate),
		stringPtr(request.CertificateUrl),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return nil, merchantaward_errors.ErrCreateMerchantAward.WithInternal(err)
	}

	return &award, nil
}

func (r *merchantAwardCommandRepository) Update(ctx context.Context, request *requests.UpdateMerchantCertificationOrAwardRequest) (*db.UpdateMerchantCertificationOrAwardRow, error) {
	var res db.UpdateMerchantCertificationOrAwardRow

	if err := r.db.GetContext(ctx, &res, updateMerchantCertificationOrAward,
		int32(*request.MerchantCertificationID),
		request.Title,
		stringPtr(request.Description),
		stringPtr(request.IssuedBy),
		parseDateToPgDate(request.IssueDate),
		parseDateToPgDate(request.ExpiryDate),
		stringPtr(request.CertificateUrl),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return nil, merchantaward_errors.ErrUpdateMerchantAward.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantAwardCommandRepository) Trash(ctx context.Context, award_id int) (*db.MerchantCertificationsAndAward, error) {
	var res db.MerchantCertificationsAndAward

	if err := r.db.GetContext(ctx, &res, trashMerchantCertificationOrAward, int32(award_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return nil, merchantaward_errors.ErrTrashedMerchantAward.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantAwardCommandRepository) Restore(ctx context.Context, award_id int) (*db.MerchantCertificationsAndAward, error) {
	var res db.MerchantCertificationsAndAward

	if err := r.db.GetContext(ctx, &res, restoreMerchantCertificationOrAward, int32(award_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return nil, merchantaward_errors.ErrRestoreMerchantAward.WithInternal(err)
	}

	return &res, nil
}

func (r *merchantAwardCommandRepository) DeletePermanent(ctx context.Context, award_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteMerchantCertificationOrAwardPermanently, int32(award_id)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return false, merchantaward_errors.ErrDeleteMerchantAwardPermanent.WithInternal(err)
	}

	return true, nil
}

func (r *merchantAwardCommandRepository) RestoreAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllMerchantCertificationsAndAwards); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return false, merchantaward_errors.ErrRestoreAllMerchantAwards.WithInternal(err)
	}
	return true, nil
}

func (r *merchantAwardCommandRepository) DeleteAll(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentMerchantCertificationsAndAwards); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, merchantaward_errors.ErrMerchantAwardNotFound
		}
		return false, merchantaward_errors.ErrDeleteAllMerchantAwardsPermanent.WithInternal(err)
	}
	return true, nil
}

func parseDateToNullTime(dateStr string) sql.NullTime {
	if dateStr == "" {
		return sql.NullTime{Valid: false}
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: t, Valid: true}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseDateToPgDate(dateStr string) pgtype.Date {
	if dateStr == "" {
		return pgtype.Date{Valid: false}
	}

	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return pgtype.Date{Valid: false}
	}

	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}
