package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-cart/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/cart_errors"
)

type cartCommandHandler struct {
	pb_cart.UnimplementedCartCommandServiceServer
	cartCommand service.CartCommandService
	logger      logger.LoggerInterface
}

func NewCartCommandHandler(cartCommand service.CartCommandService, logger logger.LoggerInterface) *cartCommandHandler {
	return &cartCommandHandler{
		cartCommand: cartCommand,
		logger:      logger,
	}
}

func (h *cartCommandHandler) Create(ctx context.Context, request *pb_cart.CreateCartRequest) (*pb_cart.ApiResponseCart, error) {
	req := &requests.CreateCartRequest{
		ProductID: int(request.GetProductId()),
		UserID:    int(request.GetUserId()),
		Quantity:  int(request.GetQuantity()),
	}

	if err := req.Validate(); err != nil {
		return nil, cart_errors.ErrGrpcValidateCreateCart
	}

	cart, err := h.cartCommand.Create(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_cart.ApiResponseCart{
		Status:  "success",
		Message: "Successfully created cart",
		Data:    mapToProtoCartResponse(cart),
	}, nil
}

// Delete implements the proto RPC (method name must match the proto rpc name).
func (h *cartCommandHandler) Delete(ctx context.Context, request *pb_cart.DeleteCartRequest) (*pb_cart.ApiResponseCartDelete, error) {
	req := &requests.DeleteCartRequest{
		CartID: int(request.GetCartId()),
		UserID: int(request.GetUserId()),
	}

	if err := req.Validate(); err != nil {
		return nil, cart_errors.ErrGrpcValidateDeleteCart
	}

	_, err := h.cartCommand.DeletePermanent(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_cart.ApiResponseCartDelete{
		Status:  "success",
		Message: "Successfully deleted cart item permanently",
	}, nil
}

// DeleteAll implements the proto RPC (method name must match the proto rpc name).
func (h *cartCommandHandler) DeleteAll(ctx context.Context, request *pb_cart.DeleteAllCartRequest) (*pb_cart.ApiResponseCartAll, error) {
	cartIDs := make([]int, len(request.GetCartIds()))
	for i, id := range request.GetCartIds() {
		cartIDs[i] = int(id)
	}

	req := &requests.DeleteAllCartRequest{
		UserID:  int(request.GetUserId()),
		CartIds: cartIDs,
	}

	if err := req.Validate(); err != nil {
		return nil, cart_errors.ErrGrpcValidateDeleteAllCart
	}

	_, err := h.cartCommand.DeleteAll(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_cart.ApiResponseCartAll{
		Status:  "success",
		Message: "Successfully deleted all cart items permanently",
	}, nil
}
