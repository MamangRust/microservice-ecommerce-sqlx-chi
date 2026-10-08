package cartapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type CartBaseResponseMapper interface {
	ToResponseCart(pbResponse *pb_cart.CartResponse) *response.CartResponse
	ToResponseCarts(pbResponse []*pb_cart.CartResponse) []*response.CartResponse
	ToApiResponseCart(pbResponse *pb_cart.ApiResponseCart) *response.ApiResponseCart
}

type CartQueryResponseMapper interface {
	CartBaseResponseMapper
	ToApiResponseCartPagination(pbResponse *pb_cart.ApiResponsePaginationCart) *response.ApiResponseCartPagination
}

type CartCommandResponseMapper interface {
	CartBaseResponseMapper
	ToApiResponseCartDelete(pbResponse *pb_cart.ApiResponseCartDelete) *response.ApiResponseCartDelete
	ToApiResponseCartAll(pbResponse *pb_cart.ApiResponseCartAll) *response.ApiResponseCartAll
}
