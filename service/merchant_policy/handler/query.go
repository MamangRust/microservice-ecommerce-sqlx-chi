package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
)

type merchantPolicyQueryHandler struct {
	pb_merchant_policy.UnimplementedMerchantPolicyQueryServiceServer
	merchantPolicyService service.MerchantPoliciesQueryService
	logger                logger.LoggerInterface
}

func NewMerchantPolicyQueryHandler(
	merchantPolicyService service.MerchantPoliciesQueryService,
	logger logger.LoggerInterface,
) pb_merchant_policy.MerchantPolicyQueryServiceServer {
	return &merchantPolicyQueryHandler{
		merchantPolicyService: merchantPolicyService,
		logger:                logger,
	}
}

func (h *merchantPolicyQueryHandler) FindAll(ctx context.Context, req *pb_merchant.FindAllMerchantRequest) (*pb_merchant_policy.ApiResponsePaginationMerchantPolicies, error) {
	merchants, total, err := h.merchantPolicyService.FindAll(ctx, &requests.FindAllMerchant{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Search:   req.GetSearch(),
	})

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return mapToPaginationResponse(merchants, total), nil
}

func (h *merchantPolicyQueryHandler) FindById(ctx context.Context, req *pb_merchant_policy.FindByIdMerchantPoliciesRequest) (*pb_merchant_policy.ApiResponseMerchantPolicies, error) {
	merchant, err := h.merchantPolicyService.FindByID(ctx, int(req.GetId()))

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return mapToSingleResponse(merchant), nil
}

func (h *merchantPolicyQueryHandler) FindByActive(ctx context.Context, req *pb_merchant.FindAllMerchantRequest) (*pb_merchant_policy.ApiResponsePaginationMerchantPoliciesDeleteAt, error) {
	merchants, total, err := h.merchantPolicyService.FindActive(ctx, &requests.FindAllMerchant{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Search:   req.GetSearch(),
	})

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return mapToPaginationDeleteAtResponse(merchants, total), nil
}

func (h *merchantPolicyQueryHandler) FindByTrashed(ctx context.Context, req *pb_merchant.FindAllMerchantRequest) (*pb_merchant_policy.ApiResponsePaginationMerchantPoliciesDeleteAt, error) {
	merchants, total, err := h.merchantPolicyService.FindTrashed(ctx, &requests.FindAllMerchant{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Search:   req.GetSearch(),
	})

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return mapToPaginationDeleteAtResponse(merchants, total), nil
}
