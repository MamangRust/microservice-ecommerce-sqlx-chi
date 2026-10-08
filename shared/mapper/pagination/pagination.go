package paginationapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/common"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

func MapPaginationMeta(s *pb_common.PaginationMeta) *response.PaginationMeta {
	if s == nil {
		return nil
	}
	return &response.PaginationMeta{
		CurrentPage:  int(s.CurrentPage),
		PageSize:     int(s.PageSize),
		TotalRecords: int(s.TotalRecords),
		TotalPages:   int(s.TotalPages),
	}
}
