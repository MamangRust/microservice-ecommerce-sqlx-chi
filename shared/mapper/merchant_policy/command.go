package merchantpolicyapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type merchantPolicyCommandResponseMapper struct{}

func NewMerchantPolicyCommandResponseMapper() MerchantPolicyCommandResponseMapper {
	return &merchantPolicyCommandResponseMapper{}
}

func (m *merchantPolicyCommandResponseMapper) ToResponseMerchantPolicy(merchant *pb_merchant_policy.MerchantPoliciesResponse) *response.MerchantPoliciesResponse {
	if merchant == nil {
		return nil
	}
	return &response.MerchantPoliciesResponse{
		ID:           int(merchant.Id),
		MerchantID:   int(merchant.MerchantId),
		PolicyType:   merchant.PolicyType,
		Title:        merchant.Title,
		Description:  merchant.Description,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
		MerchantName: merchant.MerchantName,
	}
}

func (m *merchantPolicyCommandResponseMapper) ToResponsesMerchantPolicy(merchants []*pb_merchant_policy.MerchantPoliciesResponse) []*response.MerchantPoliciesResponse {
	var mappedMerchants []*response.MerchantPoliciesResponse
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchantPolicy(merchant))
	}
	return mappedMerchants
}

func (m *merchantPolicyCommandResponseMapper) ToApiResponseMerchantPolicies(pbResponse *pb_merchant_policy.ApiResponseMerchantPolicies) *response.ApiResponseMerchantPolicies {
	return &response.ApiResponseMerchantPolicies{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantPolicy(pbResponse.Data),
	}
}

func (m *merchantPolicyCommandResponseMapper) ToResponseMerchantPolicyDeleteAt(merchant *pb_merchant_policy.MerchantPoliciesResponseDeleteAt) *response.MerchantPoliciesResponseDeleteAt {
	if merchant == nil {
		return nil
	}
	var deletedAt string
	if merchant.DeletedAt != nil {
		deletedAt = merchant.DeletedAt.Value
	}

	return &response.MerchantPoliciesResponseDeleteAt{
		ID:           int(merchant.Id),
		MerchantID:   int(merchant.MerchantId),
		PolicyType:   merchant.PolicyType,
		Title:        merchant.Title,
		Description:  merchant.Description,
		CreatedAt:    merchant.CreatedAt,
		UpdatedAt:    merchant.UpdatedAt,
		MerchantName: merchant.MerchantName,
		DeletedAt:    &deletedAt,
	}
}

func (m *merchantPolicyCommandResponseMapper) ToResponsesMerchantPolicyDeleteAt(merchants []*pb_merchant_policy.MerchantPoliciesResponseDeleteAt) []*response.MerchantPoliciesResponseDeleteAt {
	var mappedMerchants []*response.MerchantPoliciesResponseDeleteAt
	for _, merchant := range merchants {
		mappedMerchants = append(mappedMerchants, m.ToResponseMerchantPolicyDeleteAt(merchant))
	}
	return mappedMerchants
}

func (m *merchantPolicyCommandResponseMapper) ToApiResponseMerchantPoliciesDeleteAt(pbResponse *pb_merchant_policy.ApiResponseMerchantPoliciesDeleteAt) *response.ApiResponseMerchantPoliciesDeleteAt {
	return &response.ApiResponseMerchantPoliciesDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    m.ToResponseMerchantPolicyDeleteAt(pbResponse.Data),
	}
}
