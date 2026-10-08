package orderapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type OrderBaseResponseMapper interface {
	ToResponseOrder(order *pb_order.OrderResponse) *response.OrderResponse
	ToResponsesOrder(orders []*pb_order.OrderResponse) []*response.OrderResponse
	ToApiResponseOrder(pbResponse *pb_order.ApiResponseOrder) *response.ApiResponseOrder
}

type OrderQueryResponseMapper interface {
	OrderBaseResponseMapper
	ToApiResponsesOrder(pbResponse *pb_order.ApiResponsesOrder) *response.ApiResponsesOrder
	ToApiResponsePaginationOrder(pbResponse *pb_order.ApiResponsePaginationOrder) *response.ApiResponsePaginationOrder
	ToApiResponsePaginationOrderDeleteAt(pbResponse *pb_order.ApiResponsePaginationOrderDeleteAt) *response.ApiResponsePaginationOrderDeleteAt
}

type OrderCommandResponseMapper interface {
	OrderBaseResponseMapper
	ToResponseOrderDeleteAt(order *pb_order.OrderResponseDeleteAt) *response.OrderResponseDeleteAt
	ToResponsesOrderDeleteAt(orders []*pb_order.OrderResponseDeleteAt) []*response.OrderResponseDeleteAt
	ToApiResponseOrderDeleteAt(pbResponse *pb_order.ApiResponseOrderDeleteAt) *response.ApiResponseOrderDeleteAt
	ToApiResponseOrderDelete(pbResponse *pb_order.ApiResponseOrderDelete) *response.ApiResponseOrderDelete
	ToApiResponseOrderAll(pbResponse *pb_order.ApiResponseOrderAll) *response.ApiResponseOrderAll
}
