package merchantdocumenthandler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_documents"
	pbmerchantdocument "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type merchantDocumentCommandHandlerApi struct {
	client       pbmerchantdocument.MerchantDocumentCommandServiceClient
	logger       logger.LoggerInterface
	mapper       apimapper.MerchantDocumentCommandResponseMapper
	upload_image upload_image.ImageUploads
}

type merchantDocumentCommandHandleDeps struct {
	client       pbmerchantdocument.MerchantDocumentCommandServiceClient
	router       chi.Router
	logger       logger.LoggerInterface
	mapper       apimapper.MerchantDocumentCommandResponseMapper
	upload_image upload_image.ImageUploads
}

func NewMerchantDocumentCommandHandleApi(params *merchantDocumentCommandHandleDeps) *merchantDocumentCommandHandlerApi {
	handler := &merchantDocumentCommandHandlerApi{
		client:       params.client,
		logger:       params.logger,
		mapper:       params.mapper,
		upload_image: params.upload_image,
	}

	params.router.Route("/api/merchant-document-command", func(routerDoc chi.Router) {
		routerDoc.Post("/create", httpx.Handler(handler.Create))
		routerDoc.Post("/update/{id}", httpx.Handler(handler.Update))
		routerDoc.Post("/update-status/{id}", httpx.Handler(handler.UpdateStatus))
		routerDoc.Post("/trashed/{id}", httpx.Handler(handler.Trash))
		routerDoc.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerDoc.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerDoc.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerDoc.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new merchant document
// @Tags Merchant Document Command
// @Description Upload and create a new merchant document record
// @Accept mpfd
// @Produce json
// @Param merchant_id formData int true "Merchant ID"
// @Param document_type formData string true "Document type (e.g., NIB, SIUP)"
// @Param document_url formData file true "Document file"
// @Success 200 {object} response.ApiResponseMerchantDocument "Successfully created merchant document"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create merchant document"
// @Router /api/merchant-document-command/create [post]
func (h *merchantDocumentCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	formData, err := h.parseMerchantDocumentCreate(w, r)
	if err != nil {
		return err
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbmerchantdocument.CreateMerchantDocumentRequest{
		MerchantId:   int32(formData.MerchantID),
		DocumentType: formData.DocumentType,
		DocumentUrl:  formData.DocumentUrl,
	})
	if err != nil {
		return h.handleGrpcError(err, "Create")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Update merchant document
// @Tags Merchant Document Command
// @Description Update an existing merchant document record
// @Accept mpfd
// @Produce json
// @Param id path int true "Document ID"
// @Param merchant_id formData int true "Merchant ID"
// @Param document_type formData string true "Document type"
// @Param document_url formData file false "New document file"
// @Param document_url_old formData string false "Old document URL if no new file is uploaded"
// @Param status formData string true "Document status"
// @Param note formData string false "Document note"
// @Success 200 {object} response.ApiResponseMerchantDocument "Successfully updated merchant document"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update merchant document"
// @Router /api/merchant-document-command/update/{id} [post]
func (h *merchantDocumentCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	formData, err := h.parseMerchantDocumentUpdate(w, r)
	if err != nil {
		return err
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbmerchantdocument.UpdateMerchantDocumentRequest{
		DocumentId:   int32(id),
		MerchantId:   int32(formData.MerchantID),
		DocumentType: formData.DocumentType,
		DocumentUrl:  formData.DocumentUrl,
		Status:       formData.Status,
		Note:         formData.Note,
	})
	if err != nil {
		return h.handleGrpcError(err, "Update")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Update merchant document status
// @Tags Merchant Document Command
// @Description Update the status and note of a merchant document
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Param body body requests.UpdateMerchantDocumentStatusRequest true "Update status request"
// @Success 200 {object} response.ApiResponseMerchantDocument "Successfully updated document status"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update document status"
// @Router /api/merchant-document-command/update-status/{id} [post]
func (h *merchantDocumentCommandHandlerApi) UpdateStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var body requests.UpdateMerchantDocumentStatusRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.UpdateStatus(ctx, &pbmerchantdocument.UpdateMerchantDocumentStatusRequest{
		DocumentId: int32(id),
		MerchantId: int32(body.MerchantID),
		Status:     body.Status,
		Note:       body.Note,
	})
	if err != nil {
		return h.handleGrpcError(err, "UpdateStatus")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Move merchant document to trash
// @Tags Merchant Document Command
// @Description Move a merchant document record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.ApiResponseMerchantDocument "Successfully moved merchant document to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid document ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move merchant document to trash"
// @Router /api/merchant-document-command/trashed/{id} [post]
func (h *merchantDocumentCommandHandlerApi) Trash(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.Trashed(ctx, &pbmerchantdocument.TrashedMerchantDocumentRequest{DocumentId: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Trash")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Restore trashed merchant document
// @Tags Merchant Document Command
// @Description Restore a trashed merchant document record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.ApiResponseMerchantDocument "Successfully restored merchant document"
// @Failure 400 {object} errors.ErrorResponse "Invalid document ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant document"
// @Router /api/merchant-document-command/restore/{id} [post]
func (h *merchantDocumentCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.Restore(ctx, &pbmerchantdocument.RestoreMerchantDocumentRequest{DocumentId: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Restore")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Permanently delete merchant document
// @Tags Merchant Document Command
// @Description Permanently delete a merchant document record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.ApiResponseMerchantDocumentDeleteAt "Successfully deleted merchant document record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid document ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant document permanently"
// @Router /api/merchant-document-command/permanent/{id} [delete]
func (h *merchantDocumentCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeletePermanent(ctx, &pbmerchantdocument.DeleteMerchantDocumentPermanentRequest{DocumentId: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "Delete")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocumentDeleteAt(res))
}

// @Security Bearer
// @Summary Restore all trashed merchant documents
// @Tags Merchant Document Command
// @Description Restore all trashed merchant document records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantDocumentAll "Successfully restored all merchant documents"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore merchant documents"
// @Router /api/merchant-document-command/restore/all [post]
func (h *merchantDocumentCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAll(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "RestoreAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocumentAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed merchant documents
// @Tags Merchant Document Command
// @Description Permanently delete all trashed merchant document records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseMerchantDocumentAll "Successfully deleted all merchant documents permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete merchant documents permanently"
// @Router /api/merchant-document-command/permanent/all [post]
func (h *merchantDocumentCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return h.handleGrpcError(err, "DeleteAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocumentAll(res))
}

func (h *merchantDocumentCommandHandlerApi) parseMerchantDocumentCreate(w http.ResponseWriter, r *http.Request) (requests.CreateMerchantDocumentRequest, error) {
	var formData requests.CreateMerchantDocumentRequest
	var err error

	formData.MerchantID, err = strconv.Atoi(r.FormValue("merchant_id"))
	if err != nil || formData.MerchantID <= 0 {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Invalid merchant ID")
	}

	formData.DocumentType = strings.TrimSpace(r.FormValue("document_type"))
	if formData.DocumentType == "" {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Document type is required")
	}

	_, file, err := r.FormFile("document_url")
	if err != nil {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Document file is required")
	}

	uploadedPath, err := h.upload_image.ProcessImageUpload(w, "uploads/merchant_document", file, true)
	if err != nil {
		return formData, err
	}
	formData.DocumentUrl = uploadedPath

	return formData, nil
}

func (h *merchantDocumentCommandHandlerApi) parseMerchantDocumentUpdate(w http.ResponseWriter, r *http.Request) (requests.UpdateMerchantDocumentRequest, error) {
	var formData requests.UpdateMerchantDocumentRequest
	var err error

	formData.MerchantID, err = strconv.Atoi(r.FormValue("merchant_id"))
	if err != nil || formData.MerchantID <= 0 {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Invalid merchant ID")
	}

	formData.DocumentType = strings.TrimSpace(r.FormValue("document_type"))
	if formData.DocumentType == "" {
		return formData, httpx.NewHTTPError(http.StatusBadRequest, "Document type is required")
	}

	_, file, err := r.FormFile("document_url")
	if err == nil {
		uploadedPath, err := h.upload_image.ProcessImageUpload(w, "uploads/merchant_document/files", file, true)
		if err != nil {
			return formData, err
		}
		formData.DocumentUrl = uploadedPath
	} else {
		formData.DocumentUrl = r.FormValue("document_url_old")
	}

	formData.Status = r.FormValue("status")
	formData.Note = r.FormValue("note")

	return formData, nil
}

func (h *merchantDocumentCommandHandlerApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to " + operation)
	return sharedErrors.ParseGrpcError(err)
}
