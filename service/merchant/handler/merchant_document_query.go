package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
)

type merchantDocumentQueryHandler struct {
	pb_merchant_document.UnimplementedMerchantDocumentQueryServiceServer
	merchantDocumentQuery service.MerchantDocumentQueryService
	logger                logger.LoggerInterface
}

func NewMerchantDocumentQueryHandler(svc service.MerchantDocumentQueryService, logger logger.LoggerInterface) pb_merchant_document.MerchantDocumentQueryServiceServer {
	return &merchantDocumentQueryHandler{
		merchantDocumentQuery: svc,
		logger:                logger,
	}
}

func (s *merchantDocumentQueryHandler) FindAll(ctx context.Context, req *pb_merchant_document.FindAllMerchantDocumentsRequest) (*pb_merchant_document.ApiResponsePaginationMerchantDocument, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindAll(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbDocuments := make([]*pb_merchant_document.MerchantDocument, len(documents))
	for i, d := range documents {
		pbDocuments[i] = mapToProtoMerchantDocumentResponse(d)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_merchant_document.ApiResponsePaginationMerchantDocument{
		Status:     "success",
		Message:    "Successfully fetched merchant documents",
		Data:       pbDocuments,
		Pagination: paginationMeta,
	}, nil
}

func (s *merchantDocumentQueryHandler) FindById(ctx context.Context, req *pb_merchant_document.FindMerchantDocumentByIdRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentQuery.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully fetched merchant document",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentQueryHandler) FindAllActive(ctx context.Context, req *pb_merchant_document.FindAllMerchantDocumentsRequest) (*pb_merchant_document.ApiResponsePaginationMerchantDocument, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindActive(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbDocuments := make([]*pb_merchant_document.MerchantDocument, len(documents))
	for i, d := range documents {
		pbDocuments[i] = mapToProtoMerchantDocumentResponse(d)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_merchant_document.ApiResponsePaginationMerchantDocument{
		Status:     "success",
		Message:    "Successfully fetched active merchant documents",
		Data:       pbDocuments,
		Pagination: paginationMeta,
	}, nil
}

func (s *merchantDocumentQueryHandler) FindAllTrashed(ctx context.Context, req *pb_merchant_document.FindAllMerchantDocumentsRequest) (*pb_merchant_document.ApiResponsePaginationMerchantDocumentAt, error) {
	page, pageSize := normalizePage(int(req.GetPage()), int(req.GetPageSize()))
	search := req.GetSearch()

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindTrashed(ctx, &reqService)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	pbDocuments := make([]*pb_merchant_document.MerchantDocumentDeleteAt, len(documents))
	for i, d := range documents {
		pbDocuments[i] = mapToProtoMerchantDocumentResponseAt(d)
	}

	paginationMeta := createPaginationMeta(page, pageSize, *totalRecords)

	return &pb_merchant_document.ApiResponsePaginationMerchantDocumentAt{
		Status:     "success",
		Message:    "Successfully fetched trashed merchant documents",
		Data:       pbDocuments,
		Pagination: paginationMeta,
	}, nil
}
