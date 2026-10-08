package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
)

type BannerQueryHandler interface {
	pb_banner.BannerQueryServiceServer
}

type BannerCommandHandler interface {
	pb_banner.BannerCommandServiceServer
}
