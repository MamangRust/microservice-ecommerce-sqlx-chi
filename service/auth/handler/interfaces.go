package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
)

type AuthHandleGrpc interface {
	pb_auth.AuthServiceServer
	LoginUser(ctx context.Context, req *pb_auth.LoginRequest) (*pb_auth.ApiResponseLogin, error)
	RegisterUser(ctx context.Context, req *pb_auth.RegisterRequest) (*pb_auth.ApiResponseRegister, error)
}
