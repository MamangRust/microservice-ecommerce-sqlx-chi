package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/role"
	"github.com/MamangRust/microservice-ecommerce-grpc-role/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/role_errors"
)

type roleQueryHandler struct {
	pb_role.UnimplementedRoleQueryServiceServer
	roleQuery service.RoleQueryService
	logger    logger.LoggerInterface
}

func NewRoleQueryHandler(roleQuery service.RoleQueryService, logger logger.LoggerInterface) pb_role.RoleQueryServiceServer {
	return &roleQueryHandler{
		roleQuery: roleQuery,
		logger:    logger,
	}
}

func (s *roleQueryHandler) FindAllRole(ctx context.Context, req *pb_role.FindAllRoleRequest) (*pb_role.ApiResponsePaginationRole, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllRole{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb_role.RoleResponse, len(roles))
	for i, role := range roles {
		protoRoles[i] = mapToProtoRoleResponse(role)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_role.ApiResponsePaginationRole{
		Status:     "success",
		Message:    "Successfully fetched role records",
		Data:       protoRoles,
		Pagination: paginationMeta,
	}, nil
}

func (s *roleQueryHandler) FindByActive(ctx context.Context, req *pb_role.FindAllRoleRequest) (*pb_role.ApiResponsePaginationRoleDeleteAt, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllRole{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb_role.RoleResponseDeleteAt, len(roles))
	for i, role := range roles {
		protoRoles[i] = mapToProtoRoleResponseDeleteAt(role)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_role.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active roles",
		Data:       protoRoles,
		Pagination: paginationMeta,
	}, nil
}

func (s *roleQueryHandler) FindByTrashed(ctx context.Context, req *pb_role.FindAllRoleRequest) (*pb_role.ApiResponsePaginationRoleDeleteAt, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllRole{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb_role.RoleResponseDeleteAt, len(roles))
	for i, role := range roles {
		protoRoles[i] = mapToProtoRoleResponseDeleteAt(role)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_role.ApiResponsePaginationRoleDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed roles",
		Data:       protoRoles,
		Pagination: paginationMeta,
	}, nil
}

func (s *roleQueryHandler) FindByIdRole(ctx context.Context, req *pb_role.FindByIdRoleRequest) (*pb_role.ApiResponseRole, error) {
	roleID := int(req.GetRoleId())
	if roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleQuery.FindByID(ctx, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_role.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role",
		Data:    mapToProtoRoleResponse(role),
	}, nil
}

func (s *roleQueryHandler) FindByNameRole(ctx context.Context, req *pb_role.FindByNameRoleRequest) (*pb_role.ApiResponseRole, error) {
	name := req.GetName()
	if name == "" {
		return nil, role_errors.ErrGrpcRoleInvalidId // Or and appropriate error for empty name
	}

	role, err := s.roleQuery.FindByName(ctx, name)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_role.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role by name",
		Data:    mapToProtoRoleResponse(role),
	}, nil
}
