package sliderapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type SliderBaseResponseMapper interface {
	ToResponseSlider(pbResponse *pb_slider.SliderResponse) *response.SliderResponse
	ToResponsesSlider(pbResponses []*pb_slider.SliderResponse) []*response.SliderResponse
}

type SliderQueryResponseMapper interface {
	SliderBaseResponseMapper
	ToApiResponseSlider(pbResponse *pb_slider.ApiResponseSlider) *response.ApiResponseSlider
	ToApiResponsesSlider(pbResponse *pb_slider.ApiResponsesSlider) *response.ApiResponsesSlider
	ToApiResponsePaginationSlider(pbResponse *pb_slider.ApiResponsePaginationSlider) *response.ApiResponsePaginationSlider
	ToApiResponsePaginationSliderDeleteAt(pbResponse *pb_slider.ApiResponsePaginationSliderDeleteAt) *response.ApiResponsePaginationSliderDeleteAt
}

type SliderCommandResponseMapper interface {
	SliderBaseResponseMapper
	ToResponseSliderDeleteAt(pbResponse *pb_slider.SliderResponseDeleteAt) *response.SliderResponseDeleteAt
	ToResponsesSliderDeleteAt(pbResponses []*pb_slider.SliderResponseDeleteAt) []*response.SliderResponseDeleteAt
	ToApiResponseSliderDeleteAt(pbResponse *pb_slider.ApiResponseSliderDeleteAt) *response.ApiResponseSliderDeleteAt
	ToApiResponseSliderDelete(pbResponse *pb_slider.ApiResponseSliderDelete) *response.ApiResponseSliderDelete
	ToApiResponseSliderAll(pbResponse *pb_slider.ApiResponseSliderAll) *response.ApiResponseSliderAll
	ToApiResponsePaginationSliderDeleteAt(pbResponse *pb_slider.ApiResponsePaginationSliderDeleteAt) *response.ApiResponsePaginationSliderDeleteAt
}
