package carthandler

import (
	"net/http"

	cart_cache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache/cart"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	apimapper "github.com/MamangRust/microservice-ecommerce-shared/mapper/cart"
	pbcart "github.com/MamangRust/microservice-ecommerce-grpc-pb/cart"
	"github.com/go-chi/chi/v5"
)

type cartCommandHandlerApi struct {
	client pbcart.CartCommandServiceClient
	logger logger.LoggerInterface
	mapper apimapper.CartCommandResponseMapper
	cache  cart_cache.CartQueryCache
}

type cartCommandHandleDeps struct {
	client pbcart.CartCommandServiceClient
	router chi.Router
	logger logger.LoggerInterface
	mapper apimapper.CartCommandResponseMapper
	cache  cart_cache.CartQueryCache
}

func NewCartCommandHandleApi(params *cartCommandHandleDeps) *cartCommandHandlerApi {
	handler := &cartCommandHandlerApi{
		client: params.client,
		logger: params.logger,
		mapper: params.mapper,
		cache:  params.cache,
	}

	params.router.Route("/api/cart-command", func(routerCart chi.Router) {
		routerCart.Post("/create", httpx.Handler(handler.Create))
		routerCart.Delete("/delete", httpx.Handler(handler.Delete))
		routerCart.Post("/delete-all", httpx.Handler(handler.DeleteAll))

	})
	return handler
}

// @Security Bearer
// @Summary Add item to cart
// @Tags Cart Command
// @Description Add a product to the user's cart
// @Accept json
// @Produce json
// @Param request body requests.CreateCartRequest true "Cart item details"
// @Success 201 {object} response.ApiResponseCart "Successfully added to cart"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 400 {object} errors.ErrorResponse "Invalid request body"
// @Failure 500 {object} errors.ErrorResponse "Failed to add to cart"
// @Router /api/cart-command/create [post]
func (h *cartCommandHandlerApi) Create(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var body requests.CreateCartRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Create(ctx, &pbcart.CreateCartRequest{
		UserId:    int32(userID),
		ProductId: int32(body.ProductID),
		Quantity:  int32(body.Quantity),
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCarts(ctx, userID)

	return httpx.JSON(w, http.StatusCreated, h.mapper.ToApiResponseCart(res))
}

// @Security Bearer
// @Summary Remove item from cart
// @Tags Cart Command
// @Description Remove a specific item from the user's cart
// @Accept json
// @Produce json
// @Param request body requests.DeleteCartRequest true "Cart ID to delete"
// @Success 200 {object} response.ApiResponseCartDelete "Successfully removed from cart"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Failed to remove from cart"
// @Router /api/cart-command/delete [delete]
func (h *cartCommandHandlerApi) Delete(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var body requests.DeleteCartRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	res, err := h.client.Delete(ctx, &pbcart.DeleteCartRequest{
		UserId: int32(userID),
		CartId: int32(body.CartID),
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCarts(ctx, userID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCartDelete(res))
}

// @Security Bearer
// @Summary Remove multiple items from cart
// @Tags Cart Command
// @Description Remove multiple specific items from the user's cart
// @Accept json
// @Produce json
// @Param request body requests.DeleteAllCartRequest true "Cart IDs to delete"
// @Success 200 {object} response.ApiResponseCartAll "Successfully removed all items from cart"
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse "Failed to remove items from cart"
// @Router /api/cart-command/delete-all [post]
func (h *cartCommandHandlerApi) DeleteAll(w http.ResponseWriter, r *http.Request) error {
	userID, ok := httpx.Get(r, "user_id").(int)
	if !ok || userID <= 0 {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var body requests.DeleteAllCartRequest
	if err := httpx.Bind(r, &body); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}
	if err := body.Validate(); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := r.Context()
	cartIdsPb := make([]int32, len(body.CartIds))
	for i, id := range body.CartIds {
		cartIdsPb[i] = int32(id)
	}

	res, err := h.client.DeleteAll(ctx, &pbcart.DeleteAllCartRequest{
		UserId:  int32(userID),
		CartIds: cartIdsPb,
	})
	if err != nil {
		return sharedErrors.ParseGrpcError(err)
	}

	h.cache.DeleteCachedCarts(ctx, userID)

	return httpx.JSON(w, http.StatusOK, h.mapper.ToApiResponseCartAll(res))
}
