package review_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"

	reviewhandler "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler/review"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	tests "github.com/MamangRust/microservice-ecommerce-test"
	"github.com/stretchr/testify/suite"
)

type ReviewApiTestSuite struct {
	tests.BaseTestSuite
	server     chi.Router
	reviewID   int
	userID     int
	prodID     int
	merchantID int
}

func (s *ReviewApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()
	s.SetupCategoryService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupShippingAddressService()
	s.SetupOrderService()
	s.SetupReviewService()

	// Seed dependencies
	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	s.prodID = s.SeedProduct(ctx, s.merchantID, catID)

	s.server = chi.NewRouter()
	s.server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "user_id", s.userID))
			next.ServeHTTP(w, r)
		})
	})

	reviewhandler.RegisterReviewHandler(&reviewhandler.DepsReview{
		Client:        s.Conns["review"],
		Router:        s.server,
		Logger:        s.Log,
		Cache:         s.GetCacheStore(),
		Observability: s.Obs,
	})
}

func (s *ReviewApiTestSuite) TestReviewApiLifecycle() {
	// 1. Create
	reqBody := requests.CreateReviewRequest{
		ProductID: s.prodID,
		Comment:   "Test Comment",
		Rating:    5,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/review-command/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Require().Equal(http.StatusCreated, rec.Code, rec.Body.String())
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.Equal(reqBody.Comment, data["comment"])
	s.reviewID = int(data["id"].(float64))

	// 2. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/review-query", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 3. FindByProduct
	s.Require().NotZero(s.prodID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/review-query/product/%d", s.prodID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByMerchant
	s.Require().NotZero(s.merchantID)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/review-query/merchant/%d", s.merchantID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/review-query/active", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. Update
	updateBody := requests.UpdateReviewRequest{
		Comment: "Updated Comment",
		Rating:  4,
	}
	body, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/review-command/update/%d", s.reviewID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/review-command/trashed/%d", s.reviewID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/review-query/trashed", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/review-command/restore/%d", s.reviewID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. DeletePermanent
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/review-command/trashed/%d", s.reviewID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/review-command/permanent/%d", s.reviewID), nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/review-command/restore/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 12. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/review-command/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestReviewApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ReviewApiTestSuite))
}
