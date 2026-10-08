package product_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	apierror "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/apierror"
	producthandler "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/product"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ProductApiTestSuite struct {
	tests.BaseTestSuite
	server       chi.Router
	productID    int
	merchantID   int
	categoryID   int
	categoryName string
}

func (s *ProductApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()

	// Seed dependencies
	ctx := context.Background()
	userID := s.SeedUser(ctx)
	s.categoryID = s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, userID)
	s.categoryName = "Seed Category" // Default from SeedCategory if not specified

	s.server = chi.NewRouter()
	apiHandler := apierror.NewApiHandler(s.Obs, s.Log)

	producthandler.RegisterProductHandler(&producthandler.DepsProduct{
		Client:     s.Conns["product"],
		Router:     s.server,
		Logger:     s.Log,
		CacheStore: s.GetCacheStore(),
		Upload:     &tests.MockImageUpload{},
		ApiHandler: apiHandler,
	})
}

func (s *ProductApiTestSuite) TestProductApiLifecycle() {
	// 1. Create
	fields := map[string]string{
		"merchant_id":    fmt.Sprintf("%d", s.merchantID),
		"category_id":    fmt.Sprintf("%d", s.categoryID),
		"name":           "Test Product",
		"description":    "Test Description",
		"price":          "1000",
		"count_in_stock": "10",
		"brand":          "Test Brand",
		"weight":         "1",
		"rating":         "5",
		"slug_product":   "test-product",
	}
	body, contentType := s.BuildMultipartRequestBody(fields, "image", "product.jpg")
	req := httptest.NewRequest(http.MethodPost, "/api/product-command/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal("Test Product", data["name"])
	s.productID = int(data["id"].(float64))

	// 2. FindById
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product-query/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal(float64(s.productID), data["id"])

	// 3. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/product-query", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/product-query/active", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. FindByMerchant
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product-query/merchant/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. FindByCategory
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/product-query/category/%s", url.PathEscape(s.categoryName)), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. Update
	updateFields := map[string]string{
		"merchant_id":    fmt.Sprintf("%d", s.merchantID),
		"category_id":    fmt.Sprintf("%d", s.categoryID),
		"name":           "Updated Product",
		"description":    "Updated Description",
		"price":          "2000",
		"count_in_stock": "20",
		"brand":          "Updated Brand",
		"weight":         "2",
		"rating":         "4",
		"slug_product":   "updated-product",
	}
	body, contentType = s.BuildMultipartRequestBody(updateFields, "image", "updated.jpg")
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product-command/update/%d", s.productID), bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	json.Unmarshal(rec.Body.Bytes(), &res)
	data = res["data"].(map[string]interface{})
	s.Equal("Updated Product", data["name"])

	// 8. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product-command/trashed/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/product-query/trashed", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product-command/restore/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/product-command/trashed/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/product-command/permanent/%d", s.productID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 12. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/product-command/restore/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 13. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/product-command/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func (s *ProductApiTestSuite) TestProductApiNotFound() {
	// GET a non-existent product must map to 404, not 500.

	req := httptest.NewRequest(http.MethodGet, "/api/product-query/999999", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusNotFound, rec.Code, "non-existent product must be 404, got %d: %s", rec.Code, rec.Body.String())
}

func (s *ProductApiTestSuite) TestProductApiInvalidID() {
	// Non-numeric path ID must be rejected as 400.

	req := httptest.NewRequest(http.MethodGet, "/api/product-query/abc", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusBadRequest, rec.Code, "invalid ID must be 400, got %d: %s", rec.Code, rec.Body.String())
}

func TestProductApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductApiTestSuite))
}
