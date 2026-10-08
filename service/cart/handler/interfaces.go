package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
)

type CartQueryHandler interface {
	pb_cart.CartQueryServiceServer
}

type CartCommandHandler interface {
	pb_cart.CartCommandServiceServer
}
