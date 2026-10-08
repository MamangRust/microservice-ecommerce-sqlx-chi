package userapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type UserBaseResponseMapper interface {
	// Converts a single user response into an API response.
	ToApiResponseUser(pbResponse *pb_user.ApiResponseUser) *response.ApiResponseUser
}

type UserQueryResponseMapper interface {
	UserBaseResponseMapper

	// Converts paginated user records into an API response.
	ToApiResponsePaginationUser(pbResponse *pb_user.ApiResponsePaginationUser) *response.ApiResponsePaginationUser

	// Converts paginated soft-deleted users into an API response.
	ToApiResponsePaginationUserDeleteAt(pbResponse *pb_user.ApiResponsePaginationUserDeleteAt) *response.ApiResponsePaginationUserDeleteAt
}

type UserCommandResponseMapper interface {
	UserBaseResponseMapper

	// Converts a soft-deleted user response into an API response.
	ToApiResponseUserDeleteAt(pbResponse *pb_user.ApiResponseUserDeleteAt) *response.ApiResponseUserDeleteAt

	// Converts a permanently deleted user response into an API response.
	ToApiResponseUserDelete(pbResponse *pb_user.ApiResponseUserDelete) *response.ApiResponseUserDelete

	// Converts all user records into an API response.
	ToApiResponseUserAll(pbResponse *pb_user.ApiResponseUserAll) *response.ApiResponseUserAll
}
