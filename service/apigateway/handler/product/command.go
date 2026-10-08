package producthandler

import (
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	product_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/product"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/product"
	pbproduct "github.com/MamangRust/microservice-ecommerce-grpc-pb/product"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

type productCommandHandlerApi struct {
	client       pbproduct.ProductCommandServiceClient
	logger       logger.LoggerInterface
	mapper       apimapper.ProductCommandResponseMapper
	cache        product_cache.ProductCommandCache
	upload_image upload_image.ImageUploads
	errors       apierror.ApiHandler
}

type productCommandHandleDeps struct {
	client       pbproduct.ProductCommandServiceClient
	router       chi.Router
	logger       logger.LoggerInterface
	mapper       apimapper.ProductCommandResponseMapper
	cache        product_cache.ProductCommandCache
	upload_image upload_image.ImageUploads
	apiHandler   apierror.ApiHandler
}

func NewProductCommandHandleApi(params *productCommandHandleDeps) *productCommandHandlerApi {
	handler := &productCommandHandlerApi{
		client:       params.client,
		logger:       params.logger,
		mapper:       params.mapper,
		cache:        params.cache,
		upload_image: params.upload_image,
		errors:       params.apiHandler,
	}

	params.router.Route("/api/product-command", func(routerProduct chi.Router) {
		routerProduct.Post("/create", httpx.Handler(handler.Create))
		routerProduct.Post("/update/{id}", httpx.Handler(handler.Update))
		routerProduct.Post("/trashed/{id}", httpx.Handler(handler.Trashed))
		routerProduct.Post("/restore/{id}", httpx.Handler(handler.Restore))
		routerProduct.Delete("/permanent/{id}", httpx.Handler(handler.DeletePermanent))
		routerProduct.Post("/restore/all", httpx.Handler(handler.RestoreAll))
		routerProduct.Post("/permanent/all", httpx.Handler(handler.DeleteAllPermanent))

	})
	return handler
}

// @Security Bearer
// @Summary Create a new product
// @Tags Product Command
// @Description Create a new product with the provided details
// @Accept mpfd
// @Produce json
// @Param merchant_id formData int true "Merchant ID"
// @Param category_id formData int true "Category ID"
// @Param name formData string true "Product name"
// @Param description formData string true "Product description"
// @Param price formData int true "Price"
// @Param count_in_stock formData int true "Stock count"
// @Param brand formData string true "Brand"
// @Param weight formData int true "Weight"
// @Param image formData file false "Product image"
// @Success 201 {object} response.ApiResponseProduct "Successfully created product"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to create product"
// @Router /api/product-command/create [post]
func (h *productCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	merchantID, _ := strconv.Atoi(r.FormValue("merchant_id"))
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	name := r.FormValue("name")
	description := r.FormValue("description")
	price, _ := strconv.Atoi(r.FormValue("price"))
	countInStock, _ := strconv.Atoi(r.FormValue("count_in_stock"))
	brand := r.FormValue("brand")
	weight, _ := strconv.Atoi(r.FormValue("weight"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	slugProduct := r.FormValue("slug_product")

	imagePath := ""
	_, file, err := r.FormFile("image")
	if err == nil {
		path, err := h.upload_image.ProcessImageUpload(w, "uploads/products", file, false)
		if err == nil {
			imagePath = path
		}
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbproduct.CreateProductRequest{
		MerchantId:   int32(merchantID),
		CategoryId:   int32(categoryID),
		Name:         name,
		Description:  description,
		Price:        int32(price),
		CountInStock: int32(countInStock),
		Brand:        brand,
		Weight:       int32(weight),
		Rating:       int32(rating),
		SlugProduct:  slugProduct,
		ImageProduct: imagePath,
	})
	if err != nil {
		if imagePath != "" {
			h.upload_image.CleanupImageOnFailure(imagePath)
		}
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, 0)

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseProduct(res))
}

// @Security Bearer
// @Summary Update an existing product
// @Tags Product Command
// @Description Update an existing product record
// @Accept mpfd
// @Produce json
// @Param id path int true "Product ID"
// @Param merchant_id formData int false "Merchant ID"
// @Param category_id formData int false "Category ID"
// @Param name formData string false "Product name"
// @Param description formData string false "Product description"
// @Param price formData int false "Price"
// @Param count_in_stock formData int false "Stock count"
// @Param brand formData string false "Brand"
// @Param weight formData int false "Weight"
// @Param image formData file false "Product image"
// @Success 200 {object} response.ApiResponseProduct "Successfully updated product"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} errors.ErrorResponse "Failed to update product"
// @Router /api/product-command/update/{id} [post]
func (h *productCommandHandlerApi) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	merchantID, _ := strconv.Atoi(r.FormValue("merchant_id"))
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	name := r.FormValue("name")
	description := r.FormValue("description")
	price, _ := strconv.Atoi(r.FormValue("price"))
	countInStock, _ := strconv.Atoi(r.FormValue("count_in_stock"))
	brand := r.FormValue("brand")
	weight, _ := strconv.Atoi(r.FormValue("weight"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))
	slugProduct := r.FormValue("slug_product")

	imagePath := ""
	_, file, err := r.FormFile("image")
	if err == nil {
		path, err := h.upload_image.ProcessImageUpload(w, "uploads/products", file, false)
		if err == nil {
			imagePath = path
		} else {
			// Handle error properly or continue if image is optional but here we might want to fail if it's provided but invalid
		}
	}

	ctx := r.Context()
	res, err := h.client.Update(ctx, &pbproduct.UpdateProductRequest{
		ProductId:    int32(id),
		MerchantId:   int32(merchantID),
		CategoryId:   int32(categoryID),
		Name:         name,
		Description:  description,
		Price:        int32(price),
		CountInStock: int32(countInStock),
		Brand:        brand,
		Weight:       int32(weight),
		Rating:       int32(rating),
		SlugProduct:  slugProduct,
		ImageProduct: imagePath,
	})
	if err != nil {
		if imagePath != "" {
			h.upload_image.CleanupImageOnFailure(imagePath)
		}
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseProduct(res))
}

// @Security Bearer
// @Summary Move product to trash
// @Tags Product Command
// @Description Move a product record to trash by its ID
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.ApiResponseProductDeleteAt "Successfully moved product to trash"
// @Failure 400 {object} errors.ErrorResponse "Invalid product ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to move product to trash"
// @Router /api/product-command/trashed/{id} [post]
func (h *productCommandHandlerApi) Trashed(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.TrashedProduct(ctx, &pbproduct.FindByIdProductRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponsesProductDeleteAt(res))
}

// @Security Bearer
// @Summary Restore a trashed product
// @Tags Product Command
// @Description Restore a trashed product record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.ApiResponseProductDeleteAt "Successfully restored product"
// @Failure 400 {object} errors.ErrorResponse "Invalid product ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore product"
// @Router /api/product-command/restore/{id} [post]
func (h *productCommandHandlerApi) Restore(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.RestoreProduct(ctx, &pbproduct.FindByIdProductRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponsesProductDeleteAt(res))
}

// @Security Bearer
// @Summary Permanently delete a product
// @Tags Product Command
// @Description Permanently delete a product record by its ID
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} response.ApiResponseProductDelete "Successfully deleted product record permanently"
// @Failure 400 {object} errors.ErrorResponse "Invalid product ID"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete product permanently"
// @Router /api/product-command/permanent/{id} [delete]
func (h *productCommandHandlerApi) DeletePermanent(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	ctx := r.Context()
	res, err := h.client.DeleteProductPermanent(ctx, &pbproduct.FindByIdProductRequest{Id: int32(id)})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, id)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseProductDelete(res))
}

// @Security Bearer
// @Summary Restore all trashed products
// @Tags Product Command
// @Description Restore all trashed product records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseProductAll "Successfully restored all products"
// @Failure 500 {object} errors.ErrorResponse "Failed to restore products"
// @Router /api/product-command/restore/all [post]
func (h *productCommandHandlerApi) RestoreAll(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.RestoreAllProduct(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseProductAll(res))
}

// @Security Bearer
// @Summary Permanently delete all trashed products
// @Tags Product Command
// @Description Permanently delete all trashed product records
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseProductAll "Successfully deleted all products permanently"
// @Failure 500 {object} errors.ErrorResponse "Failed to delete products permanently"
// @Router /api/product-command/permanent/all [post]
func (h *productCommandHandlerApi) DeleteAllPermanent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	res, err := h.client.DeleteAllProductPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedProduct(ctx, 0)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseProductAll(res))
}
