package sliderapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type sliderQueryResponseMapper struct{}

func NewSliderQueryResponseMapper() SliderQueryResponseMapper {
	return &sliderQueryResponseMapper{}
}

func (s *sliderQueryResponseMapper) ToResponseSlider(pbResponse *pb_slider.SliderResponse) *response.SliderResponse {
	return &response.SliderResponse{
		ID:        int(pbResponse.Id),
		Name:      pbResponse.Name,
		Image:     pbResponse.Image,
		CreatedAt: pbResponse.CreatedAt,
		UpdatedAt: pbResponse.UpdatedAt,
	}
}

func (s *sliderQueryResponseMapper) ToResponsesSlider(pbResponses []*pb_slider.SliderResponse) []*response.SliderResponse {
	var sliders []*response.SliderResponse
	for _, slider := range pbResponses {
		sliders = append(sliders, s.ToResponseSlider(slider))
	}
	return sliders
}

func (s *sliderQueryResponseMapper) ToApiResponseSlider(pbResponse *pb_slider.ApiResponseSlider) *response.ApiResponseSlider {
	return &response.ApiResponseSlider{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    s.ToResponseSlider(pbResponse.Data),
	}
}

func (s *sliderQueryResponseMapper) ToApiResponsesSlider(pbResponse *pb_slider.ApiResponsesSlider) *response.ApiResponsesSlider {
	return &response.ApiResponsesSlider{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    s.ToResponsesSlider(pbResponse.Data),
	}
}

func (s *sliderQueryResponseMapper) ToApiResponsePaginationSlider(pbResponse *pb_slider.ApiResponsePaginationSlider) *response.ApiResponsePaginationSlider {
	return &response.ApiResponsePaginationSlider{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       s.ToResponsesSlider(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}

func (s *sliderQueryResponseMapper) ToApiResponsePaginationSliderDeleteAt(pbResponse *pb_slider.ApiResponsePaginationSliderDeleteAt) *response.ApiResponsePaginationSliderDeleteAt {
	return &response.ApiResponsePaginationSliderDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       s.ToResponsesSliderDeleteAt(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}

func (s *sliderQueryResponseMapper) ToResponseSliderDeleteAt(pbResponse *pb_slider.SliderResponseDeleteAt) *response.SliderResponseDeleteAt {
	var deletedAt *string
	if pbResponse.DeletedAt != nil {
		val := pbResponse.DeletedAt.Value
		deletedAt = &val
	}

	return &response.SliderResponseDeleteAt{
		ID:        int(pbResponse.Id),
		Name:      pbResponse.Name,
		Image:     pbResponse.Image,
		CreatedAt: pbResponse.CreatedAt,
		UpdatedAt: pbResponse.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func (s *sliderQueryResponseMapper) ToResponsesSliderDeleteAt(pbResponses []*pb_slider.SliderResponseDeleteAt) []*response.SliderResponseDeleteAt {
	var sliders []*response.SliderResponseDeleteAt
	for _, slider := range pbResponses {
		sliders = append(sliders, s.ToResponseSliderDeleteAt(slider))
	}
	return sliders
}
