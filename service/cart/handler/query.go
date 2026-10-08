package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-cart/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
)

type cartQueryHandler struct {
	pb_cart.UnimplementedCartQueryServiceServer
	cartQuery service.CartQueryService
	logger    logger.LoggerInterface
}

func NewCartQueryHandler(cartQuery service.CartQueryService, logger logger.LoggerInterface) *cartQueryHandler {
	return &cartQueryHandler{
		cartQuery: cartQuery,
		logger:    logger,
	}
}

func (h *cartQueryHandler) FindAll(ctx context.Context, request *pb_cart.FindAllCartRequest) (*pb_cart.ApiResponsePaginationCart, error) {
	userID := int(request.GetUserId())
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllCarts{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	cartItems, totalRecords, err := h.cartQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoCartItems := make([]*pb_cart.CartResponse, len(cartItems))
	for i, cartItem := range cartItems {
		protoCartItems[i] = mapToProtoCartResponse(cartItem)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_cart.ApiResponsePaginationCart{
		Status:     "success",
		Message:    "Successfully fetched cart items",
		Data:       protoCartItems,
		Pagination: paginationMeta,
	}, nil
}
