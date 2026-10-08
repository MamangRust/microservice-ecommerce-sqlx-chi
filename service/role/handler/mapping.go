package handler

import (
	"math"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	db "github.com/MamangRust/microservice-ecommerce-grpc-role/database/schema"
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

func mapToProtoRoleResponse(m interface{}) *pb_role.RoleResponse {
	switch v := m.(type) {
	case *db.Role:
		return &pb_role.RoleResponse{
			Id:        v.RoleID,
			Name:      v.RoleName,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
	case *db.GetRolesRow:
		return &pb_role.RoleResponse{
			Id:        v.RoleID,
			Name:      v.RoleName,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
	default:
		return nil
	}
}

func mapToProtoRoleResponseDeleteAt(m interface{}) *pb_role.RoleResponseDeleteAt {
	var res *pb_role.RoleResponseDeleteAt
	var deletedAt interface{}

	switch v := m.(type) {
	case *db.Role:
		res = &pb_role.RoleResponseDeleteAt{
			Id:        v.RoleID,
			Name:      v.RoleName,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetActiveRolesRow:
		res = &pb_role.RoleResponseDeleteAt{
			Id:        v.RoleID,
			Name:      v.RoleName,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
		}
		deletedAt = v.DeletedAt
	case *db.GetTrashedRolesRow:
		res = &pb_role.RoleResponseDeleteAt{
			Id:        v.RoleID,
			Name:      v.RoleName,
			CreatedAt: formatTimestamp(v.CreatedAt),
			UpdatedAt: formatTimestamp(v.UpdatedAt),
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
