package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-order-item/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order_item"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
)

type orderItemQueryHandler struct {
	pb_order_item.UnimplementedOrderItemQueryServiceServer
	orderItemService service.OrderItemQueryService
	logger           logger.LoggerInterface
}

func NewOrderItemQueryHandler(orderItemService service.OrderItemQueryService, logger logger.LoggerInterface) *orderItemQueryHandler {
	return &orderItemQueryHandler{
		orderItemService: orderItemService,
		logger:           logger,
	}
}

func (h *orderItemQueryHandler) FindAll(ctx context.Context, request *pb_order_item.FindAllOrderItemRequest) (*pb_order_item.ApiResponsePaginationOrderItem, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := h.orderItemService.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrderItems := make([]*pb_order_item.OrderItemResponse, len(orderItems))
	for i, item := range orderItems {
		pbOrderItems[i] = mapToProtoOrderItemResponse(item)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order_item.ApiResponsePaginationOrderItem{
		Status:     "success",
		Message:    "Successfully fetched order items",
		Data:       pbOrderItems,
		Pagination: paginationMeta,
	}, nil
}

func (h *orderItemQueryHandler) FindByActive(ctx context.Context, request *pb_order_item.FindAllOrderItemRequest) (*pb_order_item.ApiResponsePaginationOrderItemDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := h.orderItemService.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrderItems := make([]*pb_order_item.OrderItemResponseDeleteAt, len(orderItems))
	for i, item := range orderItems {
		pbOrderItems[i] = mapToProtoOrderItemResponseDeleteAt(item)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order_item.ApiResponsePaginationOrderItemDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active order items",
		Data:       pbOrderItems,
		Pagination: paginationMeta,
	}, nil
}

func (h *orderItemQueryHandler) FindByTrashed(ctx context.Context, request *pb_order_item.FindAllOrderItemRequest) (*pb_order_item.ApiResponsePaginationOrderItemDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrderItems{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orderItems, totalRecords, err := h.orderItemService.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrderItems := make([]*pb_order_item.OrderItemResponseDeleteAt, len(orderItems))
	for i, item := range orderItems {
		pbOrderItems[i] = mapToProtoOrderItemResponseDeleteAt(item)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order_item.ApiResponsePaginationOrderItemDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed order items",
		Data:       pbOrderItems,
		Pagination: paginationMeta,
	}, nil
}

func (h *orderItemQueryHandler) FindOrderItemByOrder(ctx context.Context, request *pb_order_item.FindByIdOrderItemRequest) (*pb_order_item.ApiResponsesOrderItem, error) {
	id := int(request.GetId())

	orderItems, err := h.orderItemService.FindByOrder(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrderItems := make([]*pb_order_item.OrderItemResponse, len(orderItems))
	for i, item := range orderItems {
		pbOrderItems[i] = mapToProtoOrderItemResponse(item)
	}

	return &pb_order_item.ApiResponsesOrderItem{
		Status:  "success",
		Message: "Successfully fetched order items by order",
		Data:    pbOrderItems,
	}, nil
}
