package handler

import (
	"math"

	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
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

func mapToProtoOrderResponse(m interface{}) *pb_order.OrderResponse {
	switch v := m.(type) {
	case *db.Order:
		return &pb_order.OrderResponse{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
	case *db.GetOrdersRow:
		return &pb_order.OrderResponse{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
	case *db.GetOrderByIDRow:
		return &pb_order.OrderResponse{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
	case *db.CreateOrderRow:
		return &pb_order.OrderResponse{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
	case *db.UpdateOrderRow:
		return &pb_order.OrderResponse{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
	default:
		return nil
	}
}

func mapToProtoOrderResponseDeleteAt(m interface{}) *pb_order.OrderResponseDeleteAt {
	var res *pb_order.OrderResponseDeleteAt
	var deletedAt interface{}

	switch v := m.(type) {
	case *db.Order:
		res = &pb_order.OrderResponseDeleteAt{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetOrdersActiveRow:
		res = &pb_order.OrderResponseDeleteAt{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetOrdersTrashedRow:
		res = &pb_order.OrderResponseDeleteAt{
			Id:         v.OrderID,
			MerchantId: v.MerchantID,
			UserId:     v.UserID,
			TotalPrice: int32(v.TotalPrice),
			CreatedAt:  formatTimestamp(v.CreatedAt),
			UpdatedAt:  formatTimestamp(v.UpdatedAt),
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
