package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_business"
)

type MerchantBusinessQueryHandler interface {
	pb_merchant_business.MerchantBusinessQueryServiceServer
}

type MerchantBusinessCommandHandler interface {
	pb_merchant_business.MerchantBusinessCommandServiceServer
}
