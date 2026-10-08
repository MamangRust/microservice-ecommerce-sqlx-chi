package roleapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type roleCommandResponseMapper struct {
}

func NewRoleCommandResponseMapper() RoleCommandResponseMapper {
	return &roleCommandResponseMapper{}
}

func (s *roleCommandResponseMapper) ToApiResponseRole(pbResponse *pb_role.ApiResponseRole) *response.ApiResponseRole {
	return &response.ApiResponseRole{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    s.mapResponseRole(pbResponse.Data),
	}
}

func (s *roleCommandResponseMapper) ToApiResponseRoleDelete(pbResponse *pb_role.ApiResponseRoleDelete) *response.ApiResponseRoleDelete {
	return &response.ApiResponseRoleDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *roleCommandResponseMapper) ToApiResponseRoleAll(pbResponse *pb_role.ApiResponseRoleAll) *response.ApiResponseRoleAll {
	return &response.ApiResponseRoleAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *roleCommandResponseMapper) mapResponseRole(role *pb_role.RoleResponse) *response.RoleResponse {
	if role == nil {
		return nil
	}
	return &response.RoleResponse{
		ID:        int(role.Id),
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}
