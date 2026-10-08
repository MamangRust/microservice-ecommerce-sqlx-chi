package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderQueryHandler interface {
	pb_order.OrderQueryServiceServer
}

type OrderCommandHandler interface {
	pb_order.OrderCommandServiceServer
}

type OrderHandleGrpc interface {
	FindAll(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrder, error)
	FindById(ctx context.Context, request *pb_order.FindByIdOrderRequest) (*pb_order.ApiResponseOrder, error)

	FindByActive(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrderDeleteAt, error)
	FindByTrashed(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrderDeleteAt, error)

	Create(ctx context.Context, request *pb_order.CreateOrderRequest) (*pb_order.ApiResponseOrder, error)
	Update(ctx context.Context, request *pb_order.UpdateOrderRequest) (*pb_order.ApiResponseOrder, error)
	TrashedOrder(ctx context.Context, request *pb_order.FindByIdOrderRequest) (*pb_order.ApiResponseOrderDeleteAt, error)
	RestoreOrder(ctx context.Context, request *pb_order.FindByIdOrderRequest) (*pb_order.ApiResponseOrderDeleteAt, error)
	DeleteOrderPermanent(ctx context.Context, request *pb_order.FindByIdOrderRequest) (*pb_order.ApiResponseOrderDelete, error)
	RestoreAllOrder(ctx context.Context, _ *emptypb.Empty) (*pb_order.ApiResponseOrderAll, error)
	DeleteAllOrderPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_order.ApiResponseOrderAll, error)
}
