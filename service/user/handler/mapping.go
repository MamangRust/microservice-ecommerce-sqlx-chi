package handler

import (
	"math"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	db "github.com/MamangRust/microservice-ecommerce-grpc-user/database/schema"
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

func mapToProtoUserResponse(m interface{}) *pb_user.UserResponse {
	switch v := m.(type) {
	case *db.User:
		return &pb_user.UserResponse{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	case *db.GetUsersRow:
		return &pb_user.UserResponse{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	case *db.GetUserByIDRow:
		return &pb_user.UserResponse{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	case *db.CreateUserRow:
		return &pb_user.UserResponse{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	default:
		return nil
	}
}

func mapToProtoUserResponseDeleteAt(m interface{}) *pb_user.UserResponseDeleteAt {
	switch v := m.(type) {
	case *db.User:
		return &pb_user.UserResponseDeleteAt{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
			DeletedAt: &wrapperspb.StringValue{Value: v.DeletedAt.Time.Format(time.RFC3339)},
		}
	case *db.GetUsersActiveRow:
		return &pb_user.UserResponseDeleteAt{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
			DeletedAt: &wrapperspb.StringValue{Value: v.DeletedAt.Time.Format(time.RFC3339)},
		}
	case *db.GetUserTrashedRow:
		return &pb_user.UserResponseDeleteAt{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
			DeletedAt: &wrapperspb.StringValue{Value: v.DeletedAt.Time.Format(time.RFC3339)},
		}
	case *db.TrashUserRow:
		return &pb_user.UserResponseDeleteAt{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
			DeletedAt: &wrapperspb.StringValue{Value: v.DeletedAt.Time.Format(time.RFC3339)},
		}
	case *db.RestoreUserRow:
		return &pb_user.UserResponseDeleteAt{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
			DeletedAt: &wrapperspb.StringValue{Value: v.DeletedAt.Time.Format(time.RFC3339)},
		}
	default:
		return nil
	}
}
func mapToProtoUserResponseWithPassword(m interface{}) *pb_user.UserResponseWithPassword {
	switch v := m.(type) {
	case *db.User:
		return &pb_user.UserResponseWithPassword{
			Id:        int32(v.UserID),
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			Password:  v.Password,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: v.UpdatedAt.Time.Format(time.RFC3339),
		}
	case *db.GetUserByEmailWithPasswordRow:
		return &pb_user.UserResponseWithPassword{
			Id:       int32(v.UserID),
			Email:    v.Email,
			Password: v.Password,
		}
	default:
		return nil
	}
}
