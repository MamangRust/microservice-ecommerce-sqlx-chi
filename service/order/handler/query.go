package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-order/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/order"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/order_errors"
)

type orderQueryHandler struct {
	pb_order.UnimplementedOrderQueryServiceServer
	orderQuery service.OrderQueryService
	logger     logger.LoggerInterface
}

func NewOrderQueryHandler(
	orderQuery service.OrderQueryService,
	logger logger.LoggerInterface,
) pb_order.OrderQueryServiceServer {
	return &orderQueryHandler{
		orderQuery: orderQuery,
		logger:     logger,
	}
}

func (s *orderQueryHandler) FindAll(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrder, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrder{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orders, totalRecords, err := s.orderQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrders := make([]*pb_order.OrderResponse, len(orders))
	for i, order := range orders {
		pbOrders[i] = mapToProtoOrderResponse(order)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order.ApiResponsePaginationOrder{
		Status:     "success",
		Message:    "Successfully fetched order",
		Data:       pbOrders,
		Pagination: paginationMeta,
	}, nil
}

func (s *orderQueryHandler) FindById(ctx context.Context, request *pb_order.FindByIdOrderRequest) (*pb_order.ApiResponseOrder, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, order_errors.ErrGrpcFailedInvalidId
	}

	order, err := s.orderQuery.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_order.ApiResponseOrder{
		Status:  "success",
		Message: "Successfully fetched order",
		Data:    mapToProtoOrderResponse(order),
	}, nil
}

func (s *orderQueryHandler) FindByActive(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrderDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrder{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orders, totalRecords, err := s.orderQuery.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrders := make([]*pb_order.OrderResponseDeleteAt, len(orders))
	for i, order := range orders {
		pbOrders[i] = mapToProtoOrderResponseDeleteAt(order)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order.ApiResponsePaginationOrderDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active order",
		Data:       pbOrders,
		Pagination: paginationMeta,
	}, nil
}

func (s *orderQueryHandler) FindByTrashed(ctx context.Context, request *pb_order.FindAllOrderRequest) (*pb_order.ApiResponsePaginationOrderDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllOrder{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	orders, totalRecords, err := s.orderQuery.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbOrders := make([]*pb_order.OrderResponseDeleteAt, len(orders))
	for i, order := range orders {
		pbOrders[i] = mapToProtoOrderResponseDeleteAt(order)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_order.ApiResponsePaginationOrderDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed order",
		Data:       pbOrders,
		Pagination: paginationMeta,
	}, nil
}
