package categoryhandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	category_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/category"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/category"
	pbcategory "github.com/MamangRust/microservice-ecommerce-grpc-pb/category"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type categoryCommandHandlerApi struct {
	client       pbcategory.CategoryCommandServiceClient
	logger       logger.LoggerInterface
	mapper       apimapper.CategoryCommandResponseMapper
	cache        category_cache.CategoryMencache
	upload_image upload_image.ImageUploads
	errors       apierror.ApiHandler
}

type categoryCommandHandleDeps struct {
	client       pbcategory.CategoryCommandServiceClient
	router       chi.Router
	logger       logger.LoggerInterface
	mapper       apimapper.CategoryCommandResponseMapper
	cache        category_cache.CategoryMencache
	upload_image upload_image.ImageUploads
	apiHandler   apierror.ApiHandler
}

func NewCategoryCommandHandleApi(params *categoryCommandHandleDeps) *categoryCommandHandlerApi {
	categoryCommandHandler := &categoryCommandHandlerApi{
		client:       params.client,
		logger:       params.logger,
		mapper:       params.mapper,
		cache:        params.cache,
		upload_image: params.upload_image,
		errors:       params.apiHandler,
	}

	params.router.Route("/api/category-command", func(routerCategory chi.Router) {

		routerCategory.Post("/create", httpx.Handler(categoryCommandHandler.Create))
		routerCategory.Post("/update/{id}", httpx.Handler(categoryCommandHandler.Update))
		routerCategory.Post("/trashed/{id}", httpx.Handler(categoryCommandHandler.Trashed))
		routerCategory.Post("/restore/{id}", httpx.Handler(categoryCommandHandler.Restore))
		routerCategory.Delete("/permanent/{id}", httpx.Handler(categoryCommandHandler.DeletePermanent))
		routerCategory.Post("/restore/all", httpx.Handler(categoryCommandHandler.RestoreAll))
		routerCategory.Delete("/permanent/all", httpx.Handler(categoryCommandHandler.DeleteAllPermanent))

	})
	return categoryCommandHandler
}

// @Security Bearer
// @Summary Create a new category
// @Tags Category Command
// @Description Create a new category with the provided details
// @Accept mpfd
// @Produce json
// @Param name formData string true "Category name"
// @Param description formData string true "Category description"
// @Param slug_category formData string false "Category slug"
// @Param image formData file false "Category image"
// @Success 201 {object} response.ApiResponseCategory "Successfully created category"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} errors.ErrorResponse "Failed to create category"
// @Router /api/category-command/create [post]
func (h *categoryCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	var req requests.CreateCategoryRequest
	if err := httpx.BindForm(r, &req); err != nil {
		return errors.NewBadRequestError("invalid request").WithInternal(err)
	}

	// Process the optional image upload first so ImageCategory is populated
	// before validation (ImageCategory is a required field).
	_, file, err := r.FormFile("image")
	var imageURL string
	if err == nil {
		imageURL, err = h.upload_image.ProcessImageUpload(w, "uploads/category", file, false)
		if err != nil {
			return err
		}
	}
	req.ImageCategory = imageURL

	if err := req.Validate(); err != nil {
		return errors.NewValidationError(nil)
	} // Simplified validation error

	slugCategory := ""
	if req.SlugCategory != nil {
		slugCategory = *req.SlugCategory
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbcategory.CreateCategoryRequest{
		Name: req.Name, Description: req.Description, SlugCategory: slugCategory, ImageCategory: imageURL,
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseCategory(res))
}

// @Security Bearer
// @Summary Update an existing category
// @Tags Category Command
// @Description Update an existing category record with the provided details
// @Accept mpfd
// @Produce json
// @Param id path int true "Category ID"
// @Param name formData string false "Category name"
// @Param description formData string false "Category description"
// @Param slug_category formData string false "Category slug"
// @Param image formData file false "Category image"
// @Success 200 {object} response.ApiResponseCategory "Successfully updated category"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body or validation error"
// @Failure 500 {object} errors.ErrorResponse "Failed to update category"
// @Router /api/category-command/update/{id} [post]
func (h *categoryCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	var req requests.UpdateCategoryRequest
	if err := httpx.BindForm(r, &req); err != nil {
		return errors.NewBadRequestError("invalid request").WithInternal(err)
	}

	_, file, err := r.FormFile("image")
	var imageURL string
	if err == nil {
		imageURL, err = h.upload_image.ProcessImageUpload(w, "uploads/category", file, false)
	}

	slugCategory := ""
	if req.SlugCategory != nil {
		slugCategory = *req.SlugCategory
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbcategory.UpdateCategoryRequest{
		CategoryId: int32(id), Name: req.Name, Description: req.Description, SlugCategory: slugCategory, ImageCategory: imageURL,
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCategoryCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategory(res))
}

// @Security Bearer
// @Summary Move category to trash
// @Tags Category Command
// @Description Move a category record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt "Successfully moved category to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid category ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move category to trash"
// @Router /api/category-command/trashed/{id} [post]
func (h *categoryCommandHandlerApi) Trashed(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.TrashedCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCategoryCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategoryDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed category
// @Tags Category Command
// @Description Restore a trashed category record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt "Successfully restored category"
// @Failure 400 {object} errors.ErrorResponse "Invalid category ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore category"
// @Router /api/category-command/restore/{id} [post]
func (h *categoryCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.RestoreCategory(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCategoryCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategoryDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a category
// @Tags Category Command
// @Description Permanently delete a category record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.ApiResponseCategoryDelete "Successfully deleted category record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid category ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete category permanently"
// @Router /api/category-command/permanent/{id} [delete]
func (h *categoryCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := r.Context()
	res, err := h.client.DeleteCategoryPermanent(ctx, &pbcategory.FindByIdCategoryRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCategoryCache(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategoryDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed categories
// @Tags Category Command
// @Description Restore all trashed category records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll "Successfully restored all categories"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore categories"
// @Router /api/category-command/restore/all [post]
func (h *categoryCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllCategory(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategoryAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed categories
// @Tags Category Command
// @Description Permanently delete all trashed category records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll "Successfully deleted all categories permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete categories permanently"
// @Router /api/category-command/permanent/all [delete]
func (h *categoryCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllCategoryPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCategoryAll(res))
}
