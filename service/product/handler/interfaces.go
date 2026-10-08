package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
)

type ProductQueryHandler interface {
	pb_product.ProductQueryServiceServer
}

type ProductCommandHandler interface {
	pb_product.ProductCommandServiceServer
}
