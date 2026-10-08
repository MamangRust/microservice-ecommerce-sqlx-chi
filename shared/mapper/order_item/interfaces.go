package orderitemapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type OrderItemBaseResponseMapper interface {
	ToResponseOrderItem(orderItem *pb_order_item.OrderItemResponse) *response.OrderItemResponse
	ToResponsesOrderItem(orderItems []*pb_order_item.OrderItemResponse) []*response.OrderItemResponse
}

type OrderItemQueryResponseMapper interface {
	OrderItemBaseResponseMapper
	ToApiResponseOrderItem(pbResponse *pb_order_item.ApiResponseOrderItem) *response.ApiResponseOrderItem
	ToApiResponsesOrderItem(pbResponse *pb_order_item.ApiResponsesOrderItem) *response.ApiResponsesOrderItem
	ToApiResponsePaginationOrderItem(pbResponse *pb_order_item.ApiResponsePaginationOrderItem) *response.ApiResponsePaginationOrderItem
	ToApiResponsePaginationOrderItemDeleteAt(pbResponse *pb_order_item.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt
}

type OrderItemCommandResponseMapper interface {
	OrderItemBaseResponseMapper
	ToApiResponseOrderItem(pbResponse *pb_order_item.ApiResponseOrderItem) *response.ApiResponseOrderItem
	ToResponseOrderItemDeleteAt(orderItem *pb_order_item.OrderItemResponseDeleteAt) *response.OrderItemResponseDeleteAt
	ToResponsesOrderItemDeleteAt(orderItems []*pb_order_item.OrderItemResponseDeleteAt) []*response.OrderItemResponseDeleteAt
	ToApiResponseOrderItemDelete(pbResponse *pb_order_item.ApiResponseOrderItemDelete) *response.ApiResponseOrderItemDelete
	ToApiResponseOrderItemAll(pbResponse *pb_order_item.ApiResponseOrderItemAll) *response.ApiResponseOrderItemAll
	ToApiResponsePaginationOrderItemDeleteAt(pbResponse *pb_order_item.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt
}
