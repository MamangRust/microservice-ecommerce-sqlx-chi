package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-banner/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/banner"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/banner_errors"
)

type bannerQueryHandler struct {
	pb_banner.UnimplementedBannerQueryServiceServer
	BannerQuery service.BannerQueryService
	logger      logger.LoggerInterface
}

func NewBannerQueryHandler(svc service.BannerQueryService, logger logger.LoggerInterface) BannerQueryHandler {
	return &bannerQueryHandler{
		BannerQuery: svc,
		logger:      logger,
	}
}

func (s *bannerQueryHandler) FindAll(ctx context.Context, request *pb_banner.FindAllBannerRequest) (*pb_banner.ApiResponsePaginationBanner, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllBanner{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	banners, totalRecords, err := s.BannerQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoBanners := make([]*pb_banner.BannerResponse, len(banners))
	for i, banner := range banners {
		protoBanners[i] = mapToProtoBannerResponse(banner)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_banner.ApiResponsePaginationBanner{
		Status:     "success",
		Message:    "Successfully fetched banners",
		Data:       protoBanners,
		Pagination: paginationMeta,
	}, nil
}

func (s *bannerQueryHandler) FindById(ctx context.Context, request *pb_banner.FindByIdBannerRequest) (*pb_banner.ApiResponseBanner, error) {
	id := int(request.GetId())
	if id == 0 {
		return nil, banner_errors.ErrGrpcBannerInvalidId
	}

	banner, err := s.BannerQuery.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_banner.ApiResponseBanner{
		Status:  "success",
		Message: "Successfully fetched banner",
		Data:    mapToProtoBannerResponse(banner),
	}, nil
}

func (s *bannerQueryHandler) FindByActive(ctx context.Context, request *pb_banner.FindAllBannerRequest) (*pb_banner.ApiResponsePaginationBannerDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllBanner{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	banners, totalRecords, err := s.BannerQuery.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoBanners := make([]*pb_banner.BannerResponseDeleteAt, len(banners))
	for i, banner := range banners {
		protoBanners[i] = mapToProtoBannerResponseDeleteAt(banner)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_banner.ApiResponsePaginationBannerDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active banners",
		Data:       protoBanners,
		Pagination: paginationMeta,
	}, nil
}

func (s *bannerQueryHandler) FindByTrashed(ctx context.Context, request *pb_banner.FindAllBannerRequest) (*pb_banner.ApiResponsePaginationBannerDeleteAt, error) {
	page, pageSize := normalizePage(int(request.GetPage()), int(request.GetPageSize()))
	search := request.GetSearch()

	reqService := requests.FindAllBanner{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	banners, totalRecords, err := s.BannerQuery.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoBanners := make([]*pb_banner.BannerResponseDeleteAt, len(banners))
	for i, banner := range banners {
		protoBanners[i] = mapToProtoBannerResponseDeleteAt(banner)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_banner.ApiResponsePaginationBannerDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed banners",
		Data:       protoBanners,
		Pagination: paginationMeta,
	}, nil
}
