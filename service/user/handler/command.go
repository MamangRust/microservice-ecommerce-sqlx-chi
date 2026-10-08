package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-grpc-user/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/user_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userCommandHandler struct {
	pb_user.UnimplementedUserCommandServiceServer
	UserCommand service.UserCommandService
	logger      logger.LoggerInterface
}

func NewUserCommandHandler(svc service.UserCommandService, logger logger.LoggerInterface) UserCommandHandler {
	return &userCommandHandler{
		UserCommand: svc,
		logger:      logger,
	}
}

func (s *userCommandHandler) Create(ctx context.Context, request *pb_user.CreateUserRequest) (*pb_user.ApiResponseUser, error) {
	req := &requests.CreateUserRequest{
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	if err := req.Validate(); err != nil {
		return nil, user_errors.ErrGrpcValidateCreateUser
	}

	user, err := s.UserCommand.Create(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUser{
		Status:  "success",
		Message: "Successfully created user",
		Data:    mapToProtoUserResponse(user),
	}, nil
}

func (s *userCommandHandler) Update(ctx context.Context, request *pb_user.UpdateUserRequest) (*pb_user.ApiResponseUser, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	req := &requests.UpdateUserRequest{
		UserID:          &id,
		FirstName:       request.GetFirstname(),
		LastName:        request.GetLastname(),
		Email:           request.GetEmail(),
		Password:        request.GetPassword(),
		ConfirmPassword: request.GetConfirmPassword(),
	}

	if err := req.Validate(); err != nil {
		return nil, user_errors.ErrGrpcValidateUpdateUser
	}

	user, err := s.UserCommand.Update(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user",
		Data:    mapToProtoUserResponse(user),
	}, nil
}

func (s *userCommandHandler) TrashedUser(ctx context.Context, request *pb_user.FindByIdUserRequest) (*pb_user.ApiResponseUserDeleteAt, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.UserCommand.Trash(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully trashed user",
		Data:    mapToProtoUserResponseDeleteAt(user),
	}, nil
}

func (s *userCommandHandler) RestoreUser(ctx context.Context, request *pb_user.FindByIdUserRequest) (*pb_user.ApiResponseUserDeleteAt, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.UserCommand.Restore(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUserDeleteAt{
		Status:  "success",
		Message: "Successfully restored user",
		Data:    mapToProtoUserResponseDeleteAt(user),
	}, nil
}

func (s *userCommandHandler) DeleteUserPermanent(ctx context.Context, request *pb_user.FindByIdUserRequest) (*pb_user.ApiResponseUserDelete, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	_, err := s.UserCommand.DeletePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUserDelete{
		Status:  "success",
		Message: "Successfully deleted user permanently",
	}, nil
}

func (s *userCommandHandler) RestoreAllUser(ctx context.Context, _ *emptypb.Empty) (*pb_user.ApiResponseUserAll, error) {
	_, err := s.UserCommand.RestoreAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully restored all users",
	}, nil
}

func (s *userCommandHandler) DeleteAllUserPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_user.ApiResponseUserAll, error) {
	_, err := s.UserCommand.DeleteAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUserAll{
		Status:  "success",
		Message: "Successfully deleted all users permanently",
	}, nil
}
func (s *userCommandHandler) UpdateIsVerified(ctx context.Context, request *pb_user.UpdateUserIsVerifiedRequest) (*pb_user.ApiResponseUser, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.UserCommand.UpdateIsVerified(ctx, id, request.GetIsVerified())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user verification status",
		Data:    mapToProtoUserResponse(user),
	}, nil
}

func (s *userCommandHandler) UpdatePassword(ctx context.Context, request *pb_user.UpdateUserPasswordRequest) (*pb_user.ApiResponseUser, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.UserCommand.UpdatePassword(ctx, id, request.GetPassword())
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_user.ApiResponseUser{
		Status:  "success",
		Message: "Successfully updated user password",
		Data:    mapToProtoUserResponse(user),
	}, nil
}
