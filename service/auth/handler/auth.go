package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-auth/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"go.uber.org/zap"
	"time"
)

type authHandleGrpc struct {
	pb_auth.UnimplementedAuthServiceServer
	registerService      service.RegistrationService
	loginService         service.LoginService
	passwordResetService service.PasswordResetService
	identifyService      service.IdentifyService
	logger               logger.LoggerInterface
}

func NewAuthHandleGrpc(authService *service.Service, logger logger.LoggerInterface) pb_auth.AuthServiceServer {
	return &authHandleGrpc{
		registerService:      authService.Register,
		loginService:         authService.Login,
		passwordResetService: authService.PasswordReset,
		identifyService:      authService.Identify,
		logger:               logger,
	}
}

func (s *authHandleGrpc) VerifyCode(ctx context.Context, req *pb_auth.VerifyCodeRequest) (*pb_auth.ApiResponseVerifyCode, error) {
	s.logger.Info("VerifyCode called", zap.String("code", req.Code))

	_, err := s.passwordResetService.VerifyCode(ctx, req.Code)
	if err != nil {
		s.logger.Error("VerifyCode failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("VerifyCode success", zap.String("code", req.Code))

	return &pb_auth.ApiResponseVerifyCode{
		Status:  "success",
		Message: "Verification successfully",
	}, nil
}

func (s *authHandleGrpc) ForgotPassword(ctx context.Context, req *pb_auth.ForgotPasswordRequest) (*pb_auth.ApiResponseForgotPassword, error) {
	s.logger.Info("ForgotPassword called", zap.String("email", req.Email))

	_, err := s.passwordResetService.ForgotPassword(ctx, req.Email)
	if err != nil {
		s.logger.Error("ForgotPassword failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("ForgotPassword successful", zap.Bool("success", true))

	return &pb_auth.ApiResponseForgotPassword{
		Status:  "success",
		Message: "ForgotPassword successful",
	}, nil
}

func (s *authHandleGrpc) ResetPassword(ctx context.Context, req *pb_auth.ResetPasswordRequest) (*pb_auth.ApiResponseResetPassword, error) {
	s.logger.Info("ResetPassword called", zap.String("reset_token", req.ResetToken))

	_, err := s.passwordResetService.ResetPassword(ctx, &requests.CreateResetPasswordRequest{
		ResetToken:      req.ResetToken,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	})
	if err != nil {
		s.logger.Error("ResetPassword failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("ResetPassword successful", zap.Bool("success", true))

	return &pb_auth.ApiResponseResetPassword{
		Status:  "success",
		Message: "Reset password successful",
	}, nil
}

func (s *authHandleGrpc) LoginUser(ctx context.Context, req *pb_auth.LoginRequest) (*pb_auth.ApiResponseLogin, error) {
	s.logger.Info("LoginUser called", zap.String("email", req.Email))

	request := &requests.AuthRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := s.loginService.Login(ctx, request)
	if err != nil {
		s.logger.Error("LoginUser failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("LoginUser successful", zap.Bool("success", true))

	return &pb_auth.ApiResponseLogin{
		Status:  "success",
		Message: "LoginUser successfull",
		Data: &pb_auth.TokenResponse{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken,
		},
	}, nil
}

func (s *authHandleGrpc) RefreshToken(ctx context.Context, req *pb_auth.RefreshTokenRequest) (*pb_auth.ApiResponseRefreshToken, error) {
	s.logger.Info("RefreshToken called")

	res, err := s.identifyService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		s.logger.Error("RefreshToken failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RefreshToken successful", zap.Bool("success", true))

	return &pb_auth.ApiResponseRefreshToken{
		Status:  "success",
		Message: "Refresh token successful",
		Data: &pb_auth.TokenResponse{
			AccessToken:  res.AccessToken,
			RefreshToken: req.RefreshToken,
		},
	}, nil
}

func (s *authHandleGrpc) GetMe(ctx context.Context, req *pb_auth.GetMeRequest) (*pb_auth.ApiResponseGetMe, error) {
	s.logger.Info("GetMe called")

	res, err := s.identifyService.GetMe(ctx, int(req.GetUserId()))
	if err != nil {
		s.logger.Error("GetMe failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("GetMe successful", zap.Bool("success", true))

	return &pb_auth.ApiResponseGetMe{
		Status:  "success",
		Message: "Get me successfully",
		Data: &pb_user.UserResponse{
			Id:        res.UserID,
			Firstname: res.Firstname,
			Lastname:  res.Lastname,
			Email:     res.Email,
			CreatedAt: timeString(res.CreatedAt),
			UpdatedAt: timeString(res.UpdatedAt),
		},
	}, nil
}

func (s *authHandleGrpc) RegisterUser(ctx context.Context, req *pb_auth.RegisterRequest) (*pb_auth.ApiResponseRegister, error) {
	s.logger.Info("RegisterUser called", zap.String("email", req.Email))

	request := &requests.RegisterRequest{
		FirstName:       req.Firstname,
		LastName:        req.Lastname,
		Email:           req.Email,
		Password:        req.Password,
		ConfirmPassword: req.ConfirmPassword,
	}

	res, err := s.registerService.Register(ctx, request)
	if err != nil {
		s.logger.Error("RegisterUser failed", zap.Any("error", err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RegisterUser successful", zap.Bool("success", true))
	return &pb_auth.ApiResponseRegister{
		Status:  "success",
		Message: "RegisterUser successful",
		Data: &pb_user.UserResponse{
			Id:        res.UserID,
			Firstname: res.Firstname,
			Lastname:  res.Lastname,
			Email:     res.Email,
			CreatedAt: timeString(res.CreatedAt),
			UpdatedAt: timeString(res.UpdatedAt),
		},
	}, nil
}

func timeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
