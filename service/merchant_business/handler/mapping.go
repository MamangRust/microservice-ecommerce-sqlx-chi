package handler

import (
	"math"

	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_business/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

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
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))
	return &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(totalRecords),
	}
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func formatTimestamp(v interface{}) string {
	switch t := v.(type) {
	case pgtype.Timestamptz:
		if t.Valid {
			return t.Time.Format("2006-01-02 15:04:05.000")
		}
	case pgtype.Timestamp:
		if t.Valid {
			return t.Time.Format("2006-01-02 15:04:05.000")
		}
	}
	return ""
}

func mapToProtoMerchantBusinessResponse(m interface{}) *pb_merchant_business.MerchantBusinessResponse {
	switch v := m.(type) {
	case *db.MerchantBusinessInformation:
		return &pb_merchant_business.MerchantBusinessResponse{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantsBusinessInformationRow:
		return &pb_merchant_business.MerchantBusinessResponse{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
	case *db.CreateMerchantBusinessInformationRow:
		return &pb_merchant_business.MerchantBusinessResponse{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
	case *db.UpdateMerchantBusinessInformationRow:
		return &pb_merchant_business.MerchantBusinessResponse{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
	case *db.GetMerchantBusinessInformationRow:
		return &pb_merchant_business.MerchantBusinessResponse{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
	default:
		return nil
	}
}

func mapToProtoMerchantBusinessResponseDeleteAt(m interface{}) *pb_merchant_business.MerchantBusinessResponseDeleteAt {
	var res *pb_merchant_business.MerchantBusinessResponseDeleteAt
	var deletedAt interface{}

	switch v := m.(type) {
	case *db.MerchantBusinessInformation:
		res = &pb_merchant_business.MerchantBusinessResponseDeleteAt{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetMerchantsBusinessInformationActiveRow:
		res = &pb_merchant_business.MerchantBusinessResponseDeleteAt{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetMerchantsBusinessInformationTrashedRow:
		res = &pb_merchant_business.MerchantBusinessResponseDeleteAt{
			Id:                v.MerchantBusinessInfoID,
			MerchantId:        v.MerchantID,
			BusinessType:      getString(v.BusinessType),
			TaxId:             getString(v.TaxID),
			EstablishedYear:   getInt32(v.EstablishedYear),
			NumberOfEmployees: getInt32(v.NumberOfEmployees),
			WebsiteUrl:        getString(v.WebsiteUrl),
			CreatedAt:         formatTimestamp(v.CreatedAt),
			UpdatedAt:         formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	default:
		return nil
	}

	if val := formatTimestamp(deletedAt); val != "" {
		res.DeletedAt = &wrapperspb.StringValue{Value: val}
	}

	return res
}
