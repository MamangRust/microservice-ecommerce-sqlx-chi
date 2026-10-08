package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-grpc-slider/service"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/slider_errors"
	"google.golang.org/protobuf/types/known/emptypb"
)

type sliderCommandHandler struct {
	pb_slider.UnimplementedSliderCommandServiceServer
	sliderCommand service.SliderCommandService
	logger        logger.LoggerInterface
}

func NewSliderCommandHandler(sliderCommand service.SliderCommandService, logger logger.LoggerInterface) *sliderCommandHandler {
	return &sliderCommandHandler{
		sliderCommand: sliderCommand,
		logger:        logger,
	}
}

func (s *sliderCommandHandler) Create(ctx context.Context, request *pb_slider.CreateSliderRequest) (*pb_slider.ApiResponseSlider, error) {
	req := &requests.CreateSliderRequest{
		Nama:     request.GetName(),
		FilePath: request.GetImage(),
	}

	if err := req.Validate(); err != nil {
		return nil, slider_errors.ErrGrpcValidateCreateSlider
	}

	slider, err := s.sliderCommand.Create(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSlider{
		Status:  "success",
		Message: "Successfully created slider",
		Data:    MapToSliderResponseCreateSliderRow(slider),
	}, nil
}

func (s *sliderCommandHandler) Update(ctx context.Context, request *pb_slider.UpdateSliderRequest) (*pb_slider.ApiResponseSlider, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, slider_errors.ErrGrpcInvalidID
	}

	req := &requests.UpdateSliderRequest{
		ID:       &id,
		Nama:     request.GetName(),
		FilePath: request.GetImage(),
	}

	if err := req.Validate(); err != nil {
		return nil, slider_errors.ErrGrpcValidateUpdateSlider
	}

	slider, err := s.sliderCommand.Update(ctx, req)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSlider{
		Status:  "success",
		Message: "Successfully updated slider",
		Data:    MapToSliderResponseUpdateSliderRow(slider),
	}, nil
}

func (s *sliderCommandHandler) TrashedSlider(ctx context.Context, request *pb_slider.FindByIdSliderRequest) (*pb_slider.ApiResponseSliderDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, slider_errors.ErrGrpcInvalidID
	}

	slider, err := s.sliderCommand.Trash(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSliderDeleteAt{
		Status:  "success",
		Message: "Successfully trashed slider",
		Data:    MapToSliderResponseDeleteAt(slider),
	}, nil
}

func (s *sliderCommandHandler) RestoreSlider(ctx context.Context, request *pb_slider.FindByIdSliderRequest) (*pb_slider.ApiResponseSliderDeleteAt, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, slider_errors.ErrGrpcInvalidID
	}

	slider, err := s.sliderCommand.Restore(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSliderDeleteAt{
		Status:  "success",
		Message: "Successfully restored slider",
		Data:    MapToSliderResponseDeleteAt(slider),
	}, nil
}

func (s *sliderCommandHandler) DeleteSliderPermanent(ctx context.Context, request *pb_slider.FindByIdSliderRequest) (*pb_slider.ApiResponseSliderDelete, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, slider_errors.ErrGrpcInvalidID
	}

	_, err := s.sliderCommand.DeletePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSliderDelete{
		Status:  "success",
		Message: "Successfully deleted slider permanently",
	}, nil
}

func (s *sliderCommandHandler) RestoreAllSlider(ctx context.Context, _ *emptypb.Empty) (*pb_slider.ApiResponseSliderAll, error) {
	_, err := s.sliderCommand.RestoreAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSliderAll{
		Status:  "success",
		Message: "Successfully restored all sliders",
	}, nil
}

func (s *sliderCommandHandler) DeleteAllSliderPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_slider.ApiResponseSliderAll, error) {
	_, err := s.sliderCommand.DeleteAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_slider.ApiResponseSliderAll{
		Status:  "success",
		Message: "Successfully deleted all sliders permanently",
	}, nil
}
