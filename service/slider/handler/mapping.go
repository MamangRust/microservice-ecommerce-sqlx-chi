package handler

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/slider"
	db "github.com/MamangRust/microservice-ecommerce-grpc-slider/database/schema"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func MapToSliderResponse(slider *db.Slider) *pb_slider.SliderResponse {
	return &pb_slider.SliderResponse{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
	}
}

func MapToSliderResponseGetSlidersRow(slider *db.GetSlidersRow) *pb_slider.SliderResponse {
	return &pb_slider.SliderResponse{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
	}
}

func MapToSliderResponseGetSliderByIDRow(slider *db.GetSliderByIDRow) *pb_slider.SliderResponse {
	return &pb_slider.SliderResponse{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
	}
}

func MapToSliderResponseCreateSliderRow(slider *db.CreateSliderRow) *pb_slider.SliderResponse {
	return &pb_slider.SliderResponse{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
	}
}

func MapToSliderResponseUpdateSliderRow(slider *db.UpdateSliderRow) *pb_slider.SliderResponse {
	return &pb_slider.SliderResponse{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
	}
}

func MapToSliderResponseDeleteAt(slider *db.Slider) *pb_slider.SliderResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if slider.DeletedAt.Valid {
		deletedAt = &wrapperspb.StringValue{Value: slider.DeletedAt.Time.Format("2006-01-02")}
	}

	return &pb_slider.SliderResponseDeleteAt{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
		DeletedAt: deletedAt,
	}
}

func MapToSliderResponseDeleteAtGetSlidersActiveRow(slider *db.GetSlidersActiveRow) *pb_slider.SliderResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if slider.DeletedAt.Valid {
		deletedAt = &wrapperspb.StringValue{Value: slider.DeletedAt.Time.Format("2006-01-02")}
	}

	return &pb_slider.SliderResponseDeleteAt{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
		DeletedAt: deletedAt,
	}
}

func MapToSliderResponseDeleteAtGetSlidersTrashedRow(slider *db.GetSlidersTrashedRow) *pb_slider.SliderResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if slider.DeletedAt.Valid {
		deletedAt = &wrapperspb.StringValue{Value: slider.DeletedAt.Time.Format("2006-01-02")}
	}

	return &pb_slider.SliderResponseDeleteAt{
		Id:        int32(slider.SliderID),
		Name:      slider.Name,
		Image:     slider.Image,
		CreatedAt: slider.CreatedAt.Time.Format("2006-01-02"),
		UpdatedAt: slider.UpdatedAt.Time.Format("2006-01-02"),
		DeletedAt: deletedAt,
	}
}
