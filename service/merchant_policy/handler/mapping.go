package handler

import (
	db "github.com/MamangRust/microservice-ecommerce-grpc-merchant_policy/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func mapToSingleResponse(data interface{}) *pb_merchant_policy.ApiResponseMerchantPolicies {
	return &pb_merchant_policy.ApiResponseMerchantPolicies{
		Status:  "success",
		Message: "Successfully fetched merchant policy",
		Data:    mapToMerchantPolicyResponse(data).(*pb_merchant_policy.MerchantPoliciesResponse),
	}
}

func mapToPaginationResponse(data []*db.GetMerchantPoliciesRow, total *int) *pb_merchant_policy.ApiResponsePaginationMerchantPolicies {
	var policies []*pb_merchant_policy.MerchantPoliciesResponse
	for _, v := range data {
		policies = append(policies, mapToMerchantPolicyResponse(v).(*pb_merchant_policy.MerchantPoliciesResponse))
	}

	return &pb_merchant_policy.ApiResponsePaginationMerchantPolicies{
		Status:  "success",
		Message: "Successfully fetched merchant policies",
		Data:    policies,
		Pagination: &pb_common.PaginationMeta{
			TotalRecords: int32(*total),
		},
	}
}

func mapToPaginationDeleteAtResponse(data interface{}, total *int) *pb_merchant_policy.ApiResponsePaginationMerchantPoliciesDeleteAt {
	var policies []*pb_merchant_policy.MerchantPoliciesResponseDeleteAt

	switch v := data.(type) {
	case []*db.GetMerchantPoliciesActiveRow:
		for _, item := range v {
			policies = append(policies, mapToMerchantPolicyResponse(item).(*pb_merchant_policy.MerchantPoliciesResponseDeleteAt))
		}
	case []*db.GetMerchantPoliciesTrashedRow:
		for _, item := range v {
			policies = append(policies, mapToMerchantPolicyResponse(item).(*pb_merchant_policy.MerchantPoliciesResponseDeleteAt))
		}
	}

	return &pb_merchant_policy.ApiResponsePaginationMerchantPoliciesDeleteAt{
		Status:  "success",
		Message: "Successfully fetched merchant policies",
		Data:    policies,
		Pagination: &pb_common.PaginationMeta{
			TotalRecords: int32(*total),
		},
	}
}

func mapToSingleDeleteAtResponse(data *db.MerchantPolicy) *pb_merchant_policy.ApiResponseMerchantPoliciesDeleteAt {
	return &pb_merchant_policy.ApiResponseMerchantPoliciesDeleteAt{
		Status:  "success",
		Message: "Successfully processed merchant policy",
		Data:    mapToMerchantPolicyResponse(data).(*pb_merchant_policy.MerchantPoliciesResponseDeleteAt),
	}
}

func mapToMerchantPolicyResponse(data interface{}) interface{} {
	switch v := data.(type) {
	case *db.GetMerchantPolicyRow:
		return &pb_merchant_policy.MerchantPoliciesResponse{
			Id:          int32(v.MerchantPolicyID),
			MerchantId:  int32(v.MerchantID),
			PolicyType:  v.PolicyType,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Time.String(),
			UpdatedAt:   v.UpdatedAt.Time.String(),
		}
	case *db.GetMerchantPoliciesRow:
		return &pb_merchant_policy.MerchantPoliciesResponse{
			Id:           int32(v.MerchantPolicyID),
			MerchantId:   int32(v.MerchantID),
			PolicyType:   v.PolicyType,
			Title:        v.Title,
			Description:  v.Description,
			CreatedAt:    v.CreatedAt.Time.String(),
			UpdatedAt:    v.UpdatedAt.Time.String(),
			MerchantName: v.MerchantName,
		}
	case *db.GetMerchantPoliciesActiveRow:
		return &pb_merchant_policy.MerchantPoliciesResponseDeleteAt{
			Id:           int32(v.MerchantPolicyID),
			MerchantId:   int32(v.MerchantID),
			PolicyType:   v.PolicyType,
			Title:        v.Title,
			Description:  v.Description,
			CreatedAt:    v.CreatedAt.Time.String(),
			UpdatedAt:    v.UpdatedAt.Time.String(),
			MerchantName: v.MerchantName,
			DeletedAt:    &wrapperspb.StringValue{Value: v.DeletedAt.Time.String()},
		}
	case *db.GetMerchantPoliciesTrashedRow:
		return &pb_merchant_policy.MerchantPoliciesResponseDeleteAt{
			Id:           int32(v.MerchantPolicyID),
			MerchantId:   int32(v.MerchantID),
			PolicyType:   v.PolicyType,
			Title:        v.Title,
			Description:  v.Description,
			CreatedAt:    v.CreatedAt.Time.String(),
			UpdatedAt:    v.UpdatedAt.Time.String(),
			DeletedAt:    &wrapperspb.StringValue{Value: v.DeletedAt.Time.String()},
			MerchantName: v.MerchantName,
		}
	case *db.CreateMerchantPolicyRow:
		return &pb_merchant_policy.MerchantPoliciesResponse{
			Id:          int32(v.MerchantPolicyID),
			MerchantId:  int32(v.MerchantID),
			PolicyType:  v.PolicyType,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Time.String(),
			UpdatedAt:   v.UpdatedAt.Time.String(),
		}
	case *db.UpdateMerchantPolicyRow:
		return &pb_merchant_policy.MerchantPoliciesResponse{
			Id:          int32(v.MerchantPolicyID),
			MerchantId:  int32(v.MerchantID),
			PolicyType:  v.PolicyType,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Time.String(),
			UpdatedAt:   v.UpdatedAt.Time.String(),
		}
	case *db.MerchantPolicy:
		return &pb_merchant_policy.MerchantPoliciesResponseDeleteAt{
			Id:          int32(v.MerchantPolicyID),
			MerchantId:  int32(v.MerchantID),
			PolicyType:  v.PolicyType,
			Title:       v.Title,
			Description: v.Description,
			CreatedAt:   v.CreatedAt.Time.String(),
			UpdatedAt:   v.UpdatedAt.Time.String(),
			DeletedAt:   &wrapperspb.StringValue{Value: v.DeletedAt.Time.String()},
		}
	default:
		return nil
	}
}
