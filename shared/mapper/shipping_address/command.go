package shippingaddressapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/shipping_address"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	paginationapimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/pagination"
)

type shippingAddressCommandResponseMapper struct{}

func NewShippingAddressCommandResponseMapper() ShippingAddressCommandResponseMapper {
	return &shippingAddressCommandResponseMapper{}
}

func (s *shippingAddressCommandResponseMapper) ToResponseShippingAddress(pbResponse *pb_shipping_address.ShippingResponse) *response.ShippingAddressResponse {
	return &response.ShippingAddressResponse{
		ID:             int(pbResponse.Id),
		OrderID:        int(pbResponse.OrderId),
		Alamat:         pbResponse.Alamat,
		Provinsi:       pbResponse.Provinsi,
		Negara:         pbResponse.Negara,
		Kota:           pbResponse.Kota,
		ShippingMethod: pbResponse.ShippingMethod,
		ShippingCost:   int(pbResponse.ShippingCost),
		CreatedAt:      pbResponse.CreatedAt,
		UpdatedAt:      pbResponse.UpdatedAt,
	}
}

func (s *shippingAddressCommandResponseMapper) ToResponsesShippingAddress(pbResponses []*pb_shipping_address.ShippingResponse) []*response.ShippingAddressResponse {
	var addresses []*response.ShippingAddressResponse
	for _, address := range pbResponses {
		addresses = append(addresses, s.ToResponseShippingAddress(address))
	}
	return addresses
}

func (s *shippingAddressCommandResponseMapper) ToResponseShippingAddressDeleteAt(pbResponse *pb_shipping_address.ShippingResponseDeleteAt) *response.ShippingAddressResponseDeleteAt {
	var deletedAt string
	if pbResponse.DeletedAt != nil {
		deletedAt = pbResponse.DeletedAt.Value
	}

	return &response.ShippingAddressResponseDeleteAt{
		ID:             int(pbResponse.Id),
		OrderID:        int(pbResponse.OrderId),
		Alamat:         pbResponse.Alamat,
		Provinsi:       pbResponse.Provinsi,
		Negara:         pbResponse.Negara,
		Kota:           pbResponse.Kota,
		ShippingMethod: pbResponse.ShippingMethod,
		ShippingCost:   int(pbResponse.ShippingCost),
		CreatedAt:      pbResponse.CreatedAt,
		UpdatedAt:      pbResponse.UpdatedAt,
		DeletedAt:      &deletedAt,
	}
}

func (s *shippingAddressCommandResponseMapper) ToResponsesShippingAddressDeleteAt(pbResponses []*pb_shipping_address.ShippingResponseDeleteAt) []*response.ShippingAddressResponseDeleteAt {
	var addresses []*response.ShippingAddressResponseDeleteAt
	for _, address := range pbResponses {
		addresses = append(addresses, s.ToResponseShippingAddressDeleteAt(address))
	}
	return addresses
}

func (s *shippingAddressCommandResponseMapper) ToApiResponseShippingAddressDeleteAt(pbResponse *pb_shipping_address.ApiResponseShippingDeleteAt) *response.ApiResponseShippingAddressDeleteAt {
	return &response.ApiResponseShippingAddressDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    s.ToResponseShippingAddressDeleteAt(pbResponse.Data),
	}
}

func (s *shippingAddressCommandResponseMapper) ToApiResponseShippingAddressDelete(pbResponse *pb_shipping_address.ApiResponseShippingDelete) *response.ApiResponseShippingAddressDelete {
	return &response.ApiResponseShippingAddressDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *shippingAddressCommandResponseMapper) ToApiResponseShippingAddressAll(pbResponse *pb_shipping_address.ApiResponseShippingAll) *response.ApiResponseShippingAddressAll {
	return &response.ApiResponseShippingAddressAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (s *shippingAddressCommandResponseMapper) ToApiResponsePaginationShippingAddressDeleteAt(pbResponse *pb_shipping_address.ApiResponsePaginationShippingDeleteAt) *response.ApiResponsePaginationShippingAddressDeleteAt {
	return &response.ApiResponsePaginationShippingAddressDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       s.ToResponsesShippingAddressDeleteAt(pbResponse.Data),
		Pagination: *paginationapimapper.MapPaginationMeta(pbResponse.Pagination),
	}
}
