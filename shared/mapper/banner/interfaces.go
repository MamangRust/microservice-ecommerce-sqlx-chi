package bannerapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type BannerBaseResponseMapper interface {
	ToResponseBanner(banner *pb_banner.BannerResponse) *response.BannerResponse
	ToResponsesBanner(banners []*pb_banner.BannerResponse) []*response.BannerResponse
}

type BannerQueryResponseMapper interface {
	BannerBaseResponseMapper
	ToApiResponseBanner(pbResponse *pb_banner.ApiResponseBanner) *response.ApiResponseBanner
	ToApiResponsesBanner(pbResponse *pb_banner.ApiResponsesBanner) *response.ApiResponsesBanner
	ToApiResponsePaginationBanner(pbResponse *pb_banner.ApiResponsePaginationBanner) *response.ApiResponsePaginationBanner
	ToApiResponsePaginationBannerDeleteAt(pbResponse *pb_banner.ApiResponsePaginationBannerDeleteAt) *response.ApiResponsePaginationBannerDeleteAt
}

type BannerCommandResponseMapper interface {
	BannerBaseResponseMapper
	ToApiResponseBanner(pbResponse *pb_banner.ApiResponseBanner) *response.ApiResponseBanner
	ToResponseBannerDeleteAt(banner *pb_banner.BannerResponseDeleteAt) *response.BannerResponseDeleteAt
	ToResponsesBannerDeleteAt(banners []*pb_banner.BannerResponseDeleteAt) []*response.BannerResponseDeleteAt
	ToApiResponseBannerDeleteAt(pbResponse *pb_banner.ApiResponseBannerDeleteAt) *response.ApiResponseBannerDeleteAt
	ToApiResponseBannerDelete(pbResponse *pb_banner.ApiResponseBannerDelete) *response.ApiResponseBannerDelete
	ToApiResponseBannerAll(pbResponse *pb_banner.ApiResponseBannerAll) *response.ApiResponseBannerAll
}
