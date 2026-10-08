package handler

import (
	"context"
	"math"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-slider/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
)

type sliderQueryHandler struct {
	pb_slider.UnimplementedSliderQueryServiceServer
	sliderQuery service.SliderQueryService
	logger      logger.LoggerInterface
}

func NewSliderQueryHandler(sliderQuery service.SliderQueryService, logger logger.LoggerInterface) *sliderQueryHandler {
	return &sliderQueryHandler{
		sliderQuery: sliderQuery,
		logger:      logger,
	}
}

func (s *sliderQueryHandler) FindAll(ctx context.Context, request *pb_slider.FindAllSliderRequest) (*pb_slider.ApiResponsePaginationSlider, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	sliders, totalRecords, err := s.sliderQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSliders := make([]*pb_slider.SliderResponse, len(sliders))
	for i, slider := range sliders {
		protoSliders[i] = MapToSliderResponseGetSlidersRow(slider)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_slider.ApiResponsePaginationSlider{
		Status:     "success",
		Message:    "Successfully fetched slider records",
		Data:       protoSliders,
		Pagination: paginationMeta,
	}, nil
}

func (s *sliderQueryHandler) FindByActive(ctx context.Context, request *pb_slider.FindAllSliderRequest) (*pb_slider.ApiResponsePaginationSliderDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	sliders, totalRecords, err := s.sliderQuery.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSliders := make([]*pb_slider.SliderResponseDeleteAt, len(sliders))
	for i, slider := range sliders {
		protoSliders[i] = MapToSliderResponseDeleteAtGetSlidersActiveRow(slider)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_slider.ApiResponsePaginationSliderDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active slider records",
		Data:       protoSliders,
		Pagination: paginationMeta,
	}, nil
}

func (s *sliderQueryHandler) FindByTrashed(ctx context.Context, request *pb_slider.FindAllSliderRequest) (*pb_slider.ApiResponsePaginationSliderDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllSlider{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	sliders, totalRecords, err := s.sliderQuery.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoSliders := make([]*pb_slider.SliderResponseDeleteAt, len(sliders))
	for i, slider := range sliders {
		protoSliders[i] = MapToSliderResponseDeleteAtGetSlidersTrashedRow(slider)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pb_common.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb_slider.ApiResponsePaginationSliderDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed slider records",
		Data:       protoSliders,
		Pagination: paginationMeta,
	}, nil
}

func (s *sliderQueryHandler) FindById(ctx context.Context, request *pb_slider.FindByIdSliderRequest) (*pb_slider.ApiResponseSlider, error) {
	id := int(request.GetId())

	slider, err := s.sliderQuery.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSlider{
		Status:  "success",
		Message: "Successfully fetched slider by ID",
		Data:    MapToSliderResponseGetSliderByIDRow(slider),
	}, nil
}
