package carthandler

import (
	"net/http"
	"strconv"

	cart_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/cart"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/cart"
	pbcart "github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/go-chi/chi/v5"
)

type cartQueryHandlerApi struct {
	client pbcart.CartQueryServiceClient
	logger logger.LoggerInterface
	mapper apimapper.CartQueryResponseMapper
	cache  cart_cache.CartQueryCache
}

type cartQueryHandleDeps struct {
	client pbcart.CartQueryServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.CartQueryResponseMapper
	cache  cart_cache.CartQueryCache
}

func NewCartQueryHandleApi(params *cartQueryHandleDeps) *cartQueryHandlerApi {
	handler := &cartQueryHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/cart-query", func(routerCart chi.Router) {
		routerCart.Get("/", httpx.Handler(handler.FindAll))

	})
	return handler
}

// @Security Bearer
// @Summary Find all cart items
// @Tags Cart Query
// @Description Retrieve all items in the current user's cart
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Param search query string false "Search query"
// @Success 200 {object} response.ApiResponseCartPagination "List of cart items"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Failed to retrieve cart data"
// @Router /api/cart-query [get]
func (h *cartQueryHandlerApi) FindAll(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

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
	req := &requests.FindAllCarts{UserID: userID, Page: page, PageSize: pageSize, Search: search}

	if cachedData, found := h.cache.GetCachedCarts(ctx, req); found {
		return httpx.JSON(w, http.StatusOK, cachedData)
	}

	res, err := h.client.FindAll(ctx, &pbcart.FindAllCartRequest{
		UserId:   int32(userID),
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	apiResponse := h.mapper.ToApiResponseCartPagination(res)
	h.cache.SetCachedCarts(ctx, req, apiResponse)

	return httpx.JSON(w, http.StatusOK, apiResponse)
}
