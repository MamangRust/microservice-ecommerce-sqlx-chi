package merchantdocumentsapimapper

import (
	"github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
)

type merchantDocumentCommandResponseMapper struct{}

func NewMerchantDocumentCommandResponseMapper() MerchantDocumentCommandResponseMapper {
	return &merchantDocumentCommandResponseMapper{}
}

func (m *merchantDocumentCommandResponseMapper) MapMerchantDocument(doc *pb_merchant_document.MerchantDocument) *response.MerchantDocumentResponse {
	if doc == nil {
		return nil
	}
	return &response.MerchantDocumentResponse{
		ID:           int(doc.DocumentId),
		MerchantID:   int(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		CreatedAt:    doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func (m *merchantDocumentCommandResponseMapper) MapMerchantDocuments(docs []*pb_merchant_document.MerchantDocument) []*response.MerchantDocumentResponse {
	var responses []*response.MerchantDocumentResponse
	for _, doc := range docs {
		responses = append(responses, m.MapMerchantDocument(doc))
	}
	return responses
}

func (m *merchantDocumentCommandResponseMapper) ToApiResponseMerchantDocument(doc *pb_merchant_document.ApiResponseMerchantDocument) *response.ApiResponseMerchantDocument {
	return &response.ApiResponseMerchantDocument{
		Status:  doc.Status,
		Message: doc.Message,
		Data:    m.MapMerchantDocument(doc.Data),
	}
}

func (m *merchantDocumentCommandResponseMapper) MapMerchantDocumentDeletedAt(doc *pb_merchant_document.MerchantDocumentDeleteAt) *response.MerchantDocumentResponseDeleteAt {
	if doc == nil {
		return nil
	}
	var deletedAt *string
	if doc.DeletedAt != nil {
		deletedAt = &doc.DeletedAt.Value
	}

	return &response.MerchantDocumentResponseDeleteAt{
		ID:           int(doc.DocumentId),
		MerchantID:   int(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		CreatedAt:    doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func (m *merchantDocumentCommandResponseMapper) MapMerchantDocumentsDeletedAt(docs []*pb_merchant_document.MerchantDocumentDeleteAt) []*response.MerchantDocumentResponseDeleteAt {
	var responses []*response.MerchantDocumentResponseDeleteAt
	for _, doc := range docs {
		responses = append(responses, m.MapMerchantDocumentDeletedAt(doc))
	}
	return responses
}

func (m *merchantDocumentCommandResponseMapper) ToApiResponseMerchantDocumentAll(resp *pb_merchant_document.ApiResponseMerchantDocumentAll) *response.ApiResponseMerchantDocumentAll {
	return &response.ApiResponseMerchantDocumentAll{
		Status:  resp.Status,
		Message: resp.Message,
	}
}

func (m *merchantDocumentCommandResponseMapper) ToApiResponseMerchantDocumentDeleteAt(resp *pb_merchant_document.ApiResponseMerchantDocumentDelete) *response.ApiResponseMerchantDocumentDelete {
	return &response.ApiResponseMerchantDocumentDelete{
		Status:  resp.Status,
		Message: resp.Message,
	}
}
