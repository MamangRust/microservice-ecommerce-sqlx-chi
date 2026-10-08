package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/user"
)

type UserQueryHandler interface {
	pb_user.UserQueryServiceServer
}

type UserCommandHandler interface {
	pb_user.UserCommandServiceServer
}
