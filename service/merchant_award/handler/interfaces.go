package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_award"
)

type MerchantAwardQueryHandler interface {
	pb_merchant_award.MerchantAwardQueryServiceServer
}

type MerchantAwardCommandHandler interface {
	pb_merchant_award.MerchantAwardCommandServiceServer
}
