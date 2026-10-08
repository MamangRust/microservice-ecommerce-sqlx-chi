package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_policy"
)

type MerchantPolicyQueryHandler interface {
	pb_merchant_policy.MerchantPolicyQueryServiceServer
}

type MerchantPolicyCommandHandler interface {
	pb_merchant_policy.MerchantPolicyCommandServiceServer
}
