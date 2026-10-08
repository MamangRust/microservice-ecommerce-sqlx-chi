package authapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/auth"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type AuthBaseResponseMapper interface {
}

type AuthQueryResponseMapper interface {
	AuthBaseResponseMapper
	ToResponseGetMe(res *pb_auth.ApiResponseGetMe) *response.ApiResponseGetMe
}

type AuthCommandResponseMapper interface {
	AuthBaseResponseMapper
	ToResponseVerifyCode(res *pb_auth.ApiResponseVerifyCode) *response.ApiResponseVerifyCode
	ToResponseForgotPassword(res *pb_auth.ApiResponseForgotPassword) *response.ApiResponseForgotPassword
	ToResponseResetPassword(res *pb_auth.ApiResponseResetPassword) *response.ApiResponseResetPassword
	ToResponseLogin(res *pb_auth.ApiResponseLogin) *response.ApiResponseLogin
	ToResponseRegister(res *pb_auth.ApiResponseRegister) *response.ApiResponseRegister
	ToResponseRefreshToken(res *pb_auth.ApiResponseRefreshToken) *response.ApiResponseRefreshToken
}
