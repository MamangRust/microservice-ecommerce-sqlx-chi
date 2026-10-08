package handler

import (
	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getStringValue(v interface{}) *wrapperspb.StringValue {
	if ts, ok := v.(pgtype.Timestamptz); ok && ts.Valid {
		return wrapperspb.String(ts.Time.Format("2006-01-02 15:04:05"))
	}
	if ts, ok := v.(pgtype.Timestamp); ok && ts.Valid {
		return wrapperspb.String(ts.Time.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func formatTimestamp(v interface{}) string {
	if ts, ok := v.(pgtype.Timestamptz); ok && ts.Valid {
		return ts.Time.Format("2006-01-02 15:04:05")
	}
	if ts, ok := v.(pgtype.Timestamp); ok && ts.Valid {
		return ts.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}

func createPaginationMeta(page, pageSize, totalRecords int) *pb_common.PaginationMeta {
	totalPages := (totalRecords + pageSize - 1) / pageSize
	return &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecords),
	}
}

func mapToProtoMerchantResponse(m interface{}) *pb_merchant.MerchantResponse {
	switch v := m.(type) {
	case *db.Merchant:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantsRow:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantByIDRow:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.CreateMerchantRow:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.UpdateMerchantRow:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.UpdateMerchantStatusRow:
		return &pb_merchant.MerchantResponse{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	default:
		return nil
	}
}

func mapToProtoMerchantResponseDeleteAt(m interface{}) *pb_merchant.MerchantResponseDeleteAt {
	switch v := m.(type) {
	case *db.Merchant:
		return &pb_merchant.MerchantResponseDeleteAt{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
			DeletedAt:    getStringValue(v.DeletedAt),
		}
	case *db.GetMerchantsActiveRow:
		return &pb_merchant.MerchantResponseDeleteAt{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
			DeletedAt:    getStringValue(v.DeletedAt),
		}
	default:
		return nil
	}
}

func mapToProtoMerchantResponseTrashed(m interface{}) *pb_merchant.MerchantResponseDeleteAt {
	switch v := m.(type) {
	case *db.GetMerchantsTrashedRow:
		return &pb_merchant.MerchantResponseDeleteAt{
			Id:           int32(v.MerchantID),
			UserId:       int32(v.UserID),
			Name:         v.Name,
			Description:  getString(v.Description),
			Address:      getString(v.Address),
			ContactEmail: getString(v.ContactEmail),
			ContactPhone: getString(v.ContactPhone),
			Status:       v.Status,
			CreatedAt:    formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
			DeletedAt:    getStringValue(v.DeletedAt),
		}
	default:
		return nil
	}
}

func mapToProtoMerchantDocumentResponse(m interface{}) *pb_merchant_document.MerchantDocument {
	switch v := m.(type) {
	case *db.MerchantDocument:
		return &pb_merchant_document.MerchantDocument{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantDocumentsRow:
		return &pb_merchant_document.MerchantDocument{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.GetActiveMerchantDocumentsRow:
		return &pb_merchant_document.MerchantDocument{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantDocumentRow:
		return &pb_merchant_document.MerchantDocument{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	case *db.CreateMerchantDocumentRow:
		return &pb_merchant_document.MerchantDocument{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
		}
	default:
		return nil
	}
}

func mapToProtoMerchantDocumentResponseAt(m interface{}) *pb_merchant_document.MerchantDocumentDeleteAt {
	switch v := m.(type) {
	case *db.GetTrashedMerchantDocumentsRow:
		return &pb_merchant_document.MerchantDocumentDeleteAt{
			DocumentId:   int32(v.DocumentID),
			MerchantId:   int32(v.MerchantID),
			DocumentType: v.DocumentType,
			DocumentUrl:  v.DocumentUrl,
			Status:       v.Status,
			Note:         getString(v.Note),
			UploadedAt:   formatTimestamp(v.CreatedAt),
			UpdatedAt:    formatTimestamp(v.UpdatedAt),
			DeletedAt:    getStringValue(v.DeletedAt),
		}
	default:
		return nil
	}
}
