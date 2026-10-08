package service

import (
	"context"

	models "github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type RegistrationService interface {
	Register(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
}

type LoginService interface {
	Login(ctx context.Context, request *requests.AuthRequest) (*response.TokenResponse, error)
}

type PasswordResetService interface {
	ForgotPassword(ctx context.Context, email string) (bool, error)

	ResetPassword(ctx context.Context, request *requests.CreateResetPasswordRequest) (bool, error)

	VerifyCode(ctx context.Context, code string) (bool, error)
}

type IdentifyService interface {
	RefreshToken(ctx context.Context, token string) (*response.TokenResponse, error)

	GetMe(ctx context.Context, userId int) (*models.User, error)
}
