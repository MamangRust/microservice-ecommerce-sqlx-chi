package sliderapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type sliderCommandResponseMapper struct{}

func NewSliderCommandResponseMapper() SliderCommandResponseMapper {
	return &sliderCommandResponseMapper{}
}

func (s *sliderCommandResponseMapper) ToResponseSlider(pbResponse *pb_slider.SliderResponse) *response.SliderResponse {
	return &response.SliderResponse{
		ID:        int(pbResponse.Id),
		Name:      pbResponse.Name,
		Image:     pbResponse.Image,
		CreatedAt: pbResponse.CreatedAt,
		UpdatedAt: pbResponse.UpdatedAt,
	}
}

func (s *sliderCommandResponseMapper) ToResponsesSlider(pbResponses []*pb_slider.SliderResponse) []*response.SliderResponse {
	var sliders []*response.SliderResponse
	for _, slider := range pbResponses {
		sliders = append(sliders, s.ToResponseSlider(slider))
	}
	return sliders
}

func (s *sliderCommandResponseMapper) ToResponseSliderDeleteAt(pbResponse *pb_slider.SliderResponseDeleteAt) *response.SliderResponseDeleteAt {
	var deletedAt string
	if pbResponse.DeletedAt != nil {
		deletedAt = pbResponse.DeletedAt.Value
	}

	return &response.SliderResponseDeleteAt{
		ID:        int(pbResponse.Id),
		Name:      pbResponse.Name,
		Image:     pbResponse.Image,
		CreatedAt: pbResponse.CreatedAt,
		UpdatedAt: pbResponse.UpdatedAt,
		DeletedAt: &deletedAt,
	}
}

func (s *sliderCommandResponseMapper) ToResponsesSliderDeleteAt(pbResponses []*pb_slider.SliderResponseDeleteAt) []*response.SliderResponseDeleteAt {
	var sliders []*response.SliderResponseDeleteAt
	for _, slider := range pbResponses {
		sliders = append(sliders, s.ToResponseSliderDeleteAt(slider))
	}
	return sliders
}

func (s *sliderCommandResponseMapper) ToApiResponseSliderDeleteAt(pbResponse *pb_slider.ApiResponseSliderDeleteAt) *response.ApiResponseSliderDeleteAt {
	return &response.ApiResponseSliderDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    s.ToResponseSliderDeleteAt(pbResponse.Data),
	}
}

func (s *sliderCommandResponseMapper) ToApiResponseSliderDelete(pbResponse *pb_slider.ApiResponseSliderDelete) *response.ApiResponseSliderDelete {
	return &response.ApiResponseSliderDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *sliderCommandResponseMapper) ToApiResponseSliderAll(pbResponse *pb_slider.ApiResponseSliderAll) *response.ApiResponseSliderAll {
	return &response.ApiResponseSliderAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *sliderCommandResponseMapper) ToApiResponsePaginationSliderDeleteAt(pbResponse *pb_slider.ApiResponsePaginationSliderDeleteAt) *response.ApiResponsePaginationSliderDeleteAt {
	return &response.ApiResponsePaginationSliderDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       s.ToResponsesSliderDeleteAt(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}
