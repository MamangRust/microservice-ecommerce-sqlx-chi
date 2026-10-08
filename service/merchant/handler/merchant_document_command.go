package handler

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-grpc-merchant/service"
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	merchant_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/merchant"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantDocumentCommandHandler struct {
	pb_merchant_document.UnimplementedMerchantDocumentCommandServiceServer
	merchantDocumentCommand service.MerchantDocumentCommandService
	logger                  logger.LoggerInterface
}

func NewMerchantDocumentCommandHandler(svc service.MerchantDocumentCommandService, logger logger.LoggerInterface) pb_merchant_document.MerchantDocumentCommandServiceServer {
	return &merchantDocumentCommandHandler{
		merchantDocumentCommand: svc,
		logger:                  logger,
	}
}

func (s *merchantDocumentCommandHandler) Create(ctx context.Context, req *pb_merchant_document.CreateMerchantDocumentRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	request := requests.CreateMerchantDocumentRequest{
		MerchantID:   int(req.GetMerchantId()),
		DocumentType: req.GetDocumentType(),
		DocumentUrl:  req.GetDocumentUrl(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcValidateCreateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.Create(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully created merchant document",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentCommandHandler) Update(ctx context.Context, req *pb_merchant_document.UpdateMerchantDocumentRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	request := requests.UpdateMerchantDocumentRequest{
		DocumentID:   &id,
		MerchantID:   int(req.GetMerchantId()),
		DocumentType: req.GetDocumentType(),
		DocumentUrl:  req.GetDocumentUrl(),
		Status:       req.GetStatus(),
		Note:         req.GetNote(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcFailedUpdateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.Update(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully updated merchant document",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentCommandHandler) UpdateStatus(ctx context.Context, req *pb_merchant_document.UpdateMerchantDocumentStatusRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	request := requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: &id,
		MerchantID: int(req.GetMerchantId()),
		Status:     req.GetStatus(),
		Note:       req.GetNote(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchant_errors.ErrGrpcFailedUpdateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.UpdateStatus(ctx, &request)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully updated merchant document status",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentCommandHandler) Trashed(ctx context.Context, req *pb_merchant_document.TrashedMerchantDocumentRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentCommand.Trash(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully trashed merchant document",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentCommandHandler) Restore(ctx context.Context, req *pb_merchant_document.RestoreMerchantDocumentRequest) (*pb_merchant_document.ApiResponseMerchantDocument, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentCommand.Restore(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully restored merchant document",
		Data:    mapToProtoMerchantDocumentResponse(document),
	}, nil
}

func (s *merchantDocumentCommandHandler) DeletePermanent(ctx context.Context, req *pb_merchant_document.DeleteMerchantDocumentPermanentRequest) (*pb_merchant_document.ApiResponseMerchantDocumentDelete, error) {
	id := int(req.GetDocumentId())
	if id == 0 {
		return nil, merchant_errors.ErrGrpcMerchantInvalidID
	}

	_, err := s.merchantDocumentCommand.DeletePermanent(ctx, id)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocumentDelete{
		Status:  "success",
		Message: "Successfully permanently deleted merchant document",
	}, nil
}

func (s *merchantDocumentCommandHandler) RestoreAll(ctx context.Context, _ *emptypb.Empty) (*pb_merchant_document.ApiResponseMerchantDocumentAll, error) {
	_, err := s.merchantDocumentCommand.RestoreAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocumentAll{
		Status:  "success",
		Message: "Successfully restored all merchant documents",
	}, nil
}

func (s *merchantDocumentCommandHandler) DeleteAllPermanent(ctx context.Context, _ *emptypb.Empty) (*pb_merchant_document.ApiResponseMerchantDocumentAll, error) {
	_, err := s.merchantDocumentCommand.DeleteAll(ctx)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb_merchant_document.ApiResponseMerchantDocumentAll{
		Status:  "success",
		Message: "Successfully permanently deleted all merchant documents",
	}, nil
}
