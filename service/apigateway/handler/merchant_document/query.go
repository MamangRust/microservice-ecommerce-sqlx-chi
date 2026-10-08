package merchantdocumenthandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/merchant_documents"
	pbmerchantdocument "github.com/MamangRust/microservice-ecommerce-grpc-pb/merchant_document"
	"github.com/go-chi/chi/v5"
)

type merchantDocumentQueryHandlerApi struct {
	client pbmerchantdocument.MerchantDocumentQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.MerchantDocumentQueryResponseMapper
}

type merchantDocumentQueryHandleDeps struct {
	client pbmerchantdocument.MerchantDocumentQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.MerchantDocumentQueryResponseMapper
}

func NewMerchantDocumentQueryHandleApi(params *merchantDocumentQueryHandleDeps) *merchantDocumentQueryHandlerApi {
	handler := &merchantDocumentQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
	}

	params.router.Route("/api/merchant-document-query", func(routerDoc chi.Router) {
		routerDoc.Get("/", httpx.Handler(handler.FindAll))
		routerDoc.Get("/{id}", httpx.Handler(handler.FindById))
		routerDoc.Get("/active", httpx.Handler(handler.FindByActive))
		routerDoc.Get("/trashed", httpx.Handler(handler.FindByTrashed))

	})
	return handler
}

// @Security Bearer
// @Summary Find all merchant documents
// @Tags Merchant Document Query
// @Description Retrieve a list of all merchant documents
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDocument "List of merchant documents"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant document data"
// @Router /api/merchant-document-query [get]
func (h *merchantDocumentQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	res, err := h.client.FindAll(ctx, &pbmerchantdocument.FindAllMerchantDocumentsRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindAll")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponsePaginationMerchantDocument(res))
}

// @Security Bearer
// @Summary Find merchant document by ID
// @Tags Merchant Document Query
// @Description Retrieve a merchant document by ID
// @Accept json
// @Produce json
// @Param id path int true "Document ID"
// @Success 200 {object} response.ApiResponseMerchantDocument "Merchant document data"
// @Failure 400 {object} errors.ErrorResponse "Invalid document ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant document data"
// @Router /api/merchant-document-query/{id} [get]
func (h *merchantDocumentQueryHandlerApi) FindById(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.FindById(ctx, &pbmerchantdocument.FindMerchantDocumentByIdRequest{DocumentId: int32(id)})
	if err != nil {
		return h.handleGrpcError(err, "FindById")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseMerchantDocument(res))
}

// @Security Bearer
// @Summary Retrieve active merchant documents
// @Tags Merchant Document Query
// @Description Retrieve a list of active merchant documents
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDocument "List of active merchant documents"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant document data"
// @Router /api/merchant-document-query/active [get]
func (h *merchantDocumentQueryHandlerApi) FindByActive(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	res, err := h.client.FindAllActive(ctx, &pbmerchantdocument.FindAllMerchantDocumentsRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindByActive")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponsePaginationMerchantDocument(res))
}

// @Security Bearer
// @Summary Retrieve trashed merchant documents
// @Tags Merchant Document Query
// @Description Retrieve a list of trashed merchant document records
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponsePaginationMerchantDocumentDeleteAt "List of trashed merchant document data"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve merchant document data"
// @Router /api/merchant-document-query/trashed [get]
func (h *merchantDocumentQueryHandlerApi) FindByTrashed(w http.ResponseWriter, r *http.Request) error {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 10
	}
	search := r.URL.Query().Get("search")

	ctx := r.Context()
	res, err := h.client.FindAllTrashed(ctx, &pbmerchantdocument.FindAllMerchantDocumentsRequest{
		Page: int32(page), PageSize: int32(pageSize), Search: search,
	})
	if err != nil {
		return h.handleGrpcError(err, "FindByTrashed")
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponsePaginationMerchantDocumentDeleteAt(res))
}

func (h *merchantDocumentQueryHandlerApi) handleGrpcError(err error, operation string) error {
	h.logger.Error("Failed to " + operation)
	return sharedErrors.ParseGrpcError(err)
}
