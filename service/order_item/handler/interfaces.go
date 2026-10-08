package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
)

type OrderItemQueryHandler interface {
	pb_order_item.OrderItemQueryServiceServer
}

type OrderItemCommandHandler interface {
	pb_order_item.OrderItemCommandServiceServer
}
