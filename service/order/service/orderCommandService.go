package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/MamangRust/microservice-ecommerce-grpc-order/cache"
	db "github.com/MamangRust/microservice-ecommerce-grpc-order/database/schema"
	"github.com/MamangRust/microservice-ecommerce-grpc-order/repository"
	shippingadapter "github.com/MamangRust/microservice-ecommerce-pkg/adapter/shipping_address"
	models "github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/events"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/errorhandler"
	sharedErrors "github.com/MamangRust/microservice-ecommerce-shared/errors"
	"github.com/MamangRust/microservice-ecommerce-shared/errors/order_errors"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type orderCommandService struct {
	observability              observability.TraceLoggerObservability
	cache                      cache.OrderCommandCache
	userQueryRepository        repository.UserQueryRepository
	productQueryRepository     repository.ProductQueryRepository
	productCommandRepository   repository.ProductCommandRepository
	orderQueryRepository       repository.OrderQueryRepository
	orderCommandRepository     repository.OrderCommandRepository
	orderItemQueryRepository   repository.OrderItemQueryRepository
	orderItemCommandRepos      repository.OrderItemCommandRepository
	merchantQueryRepository    repository.MerchantQueryRepository
	shippingAddressRepository  repository.ShippingAddressCommandRepository
	transactionCommandRepos    repository.TransactionCommandRepository
	shippingQueryRepository    shippingadapter.QueryRepository
	stockReservationRepository repository.StockReservationRepository
	outbox                     repository.OutboxRepository
	logger                     logger.LoggerInterface
}

type OrderCommandServiceDeps struct {
	Observability                observability.TraceLoggerObservability
	Cache                        cache.OrderCommandCache
	UserQueryRepository          repository.UserQueryRepository
	ProductQueryRepository       repository.ProductQueryRepository
	ProductCommandRepository     repository.ProductCommandRepository
	OrderQueryRepository         repository.OrderQueryRepository
	OrderCommandRepository       repository.OrderCommandRepository
	OrderItemQueryRepository     repository.OrderItemQueryRepository
	OrderItemCommandRepository   repository.OrderItemCommandRepository
	MerchantQueryRepository      repository.MerchantQueryRepository
	ShippingAddressRepository    repository.ShippingAddressCommandRepository
	TransactionCommandRepository repository.TransactionCommandRepository
	ShippingQueryRepository      shippingadapter.QueryRepository
	StockReservationRepository   repository.StockReservationRepository
	Outbox                       repository.OutboxRepository
	Logger                       logger.LoggerInterface
}

func NewOrderCommandService(deps *OrderCommandServiceDeps) OrderCommandService {
	return &orderCommandService{
		observability:              deps.Observability,
		cache:                      deps.Cache,
		userQueryRepository:        deps.UserQueryRepository,
		productQueryRepository:     deps.ProductQueryRepository,
		productCommandRepository:   deps.ProductCommandRepository,
		orderQueryRepository:       deps.OrderQueryRepository,
		orderCommandRepository:     deps.OrderCommandRepository,
		orderItemQueryRepository:   deps.OrderItemQueryRepository,
		orderItemCommandRepos:      deps.OrderItemCommandRepository,
		merchantQueryRepository:    deps.MerchantQueryRepository,
		shippingAddressRepository:  deps.ShippingAddressRepository,
		transactionCommandRepos:    deps.TransactionCommandRepository,
		shippingQueryRepository:    deps.ShippingQueryRepository,
		stockReservationRepository: deps.StockReservationRepository,
		outbox:                     deps.Outbox,
		logger:                     deps.Logger,
	}
}

func (s *orderCommandService) Create(ctx context.Context, req *requests.CreateOrderRequest) (*db.CreateOrderRow, error) {
	const method = "Create"
	if req == nil {
		return nil, sharedErrors.ErrBadRequest.WithMessage("order request is required")
	}

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("merchant.id", req.MerchantID),
		attribute.Int("user.id", req.UserID))

	defer func() {
		end(status)
	}()

	order, err := s.orderCommandRepository.Create(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: req.MerchantID,
		UserID:     req.UserID,
	})
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*db.CreateOrderRow](s.logger, err, method, span)
	}

	// The order flow spans several services. Keep a compensation list so a
	// failure after stock reservation cannot leave inventory or child records
	// behind. The SQL stock adjustment itself remains atomic per product.
	type stockAdjustment struct {
		productID int
		quantity  int
	}
	adjusted := make([]stockAdjustment, 0, len(req.Items))
	rollback := func() {
		for i := len(adjusted) - 1; i >= 0; i-- {
			reservation, reservationErr := s.stockReservationRepository.Release(ctx, int(order.OrderID), adjusted[i].productID)
			if reservationErr != nil && !errors.Is(reservationErr, sql.ErrNoRows) {
				s.logger.Error("failed to release compensated stock reservation", zap.Error(reservationErr), zap.Int("product_id", adjusted[i].productID))
				continue
			}
			if reservation != nil || errors.Is(reservationErr, sql.ErrNoRows) {
				op := fmt.Sprintf("order-create-rollback-%d-%d", order.OrderID, adjusted[i].productID)
				if _, rollbackErr := s.productCommandRepository.AdjustProductStock(ctx, adjusted[i].productID, adjusted[i].quantity, op); rollbackErr != nil {
					s.logger.Error("failed to compensate reserved product stock", zap.Error(rollbackErr), zap.Int("product_id", adjusted[i].productID), zap.Int("quantity", adjusted[i].quantity))
				}
			}
		}
		// Trash first (sets deleted_at), then the single atomic purge removes the
		// reservation ledger and all child rows together with the order.
		if _, cleanupErr := s.orderCommandRepository.Trash(ctx, int(order.OrderID)); cleanupErr != nil {
			s.logger.Error("failed to trash incomplete order", zap.Error(cleanupErr), zap.Int32("order_id", order.OrderID))
		} else if _, cleanupErr = s.orderCommandRepository.DeletePermanentWithChildren(ctx, int(order.OrderID)); cleanupErr != nil {
			s.logger.Error("failed to permanently delete incomplete order", zap.Error(cleanupErr), zap.Int32("order_id", order.OrderID))
		}
	}
	fail := func(err error) (*db.CreateOrderRow, error) {
		status = "error"
		rollback()
		return errorhandler.HandleError[*db.CreateOrderRow](s.logger, err, method, span)
	}

	for _, item := range req.Items {
		product, err := s.productQueryRepository.FindByID(ctx, item.ProductID)
		if err != nil {
			return fail(err)
		}

		// Reserve first through the atomic delta RPC. A stale read cannot cause
		// an oversell; the database guard is the final authority.
		operationID := fmt.Sprintf("order-create-%d-item-%d-product-%d-quantity-%d", order.OrderID, len(adjusted), product.ProductID, item.Quantity)
		_, err = s.productCommandRepository.AdjustProductStock(ctx, int(product.ProductID), -item.Quantity, operationID)
		if err != nil {
			// Preserve the business-level stock error when the atomic guarded
			// UPDATE returns no row because another request consumed the stock.
			currentProduct, lookupErr := s.productQueryRepository.FindByID(ctx, item.ProductID)
			if lookupErr == nil && currentProduct.CountInStock < int32(item.Quantity) {
				return fail(order_errors.ErrInsufficientProductStock)
			}
			return fail(err)
		}
		adjusted = append(adjusted, stockAdjustment{productID: int(product.ProductID), quantity: item.Quantity})
		if _, err = s.stockReservationRepository.Upsert(ctx, int(order.OrderID), int(product.ProductID), item.Quantity); err != nil {
			return fail(err)
		}

		createdItem, err := s.orderItemCommandRepos.Create(ctx, &requests.CreateOrderItemRecordRequest{
			OrderID:   int(order.OrderID),
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     int(product.Price),
		})
		if err != nil {
			return fail(err)
		}

		// Stats pipeline (F3): publish a stats.ecommerce.order_item.event so
		// stats-writer materializes this line item into ClickHouse for the
		// category pricing statistics. The envelope carries event_id for
		// idempotent consumption and the business event_time for
		// ReplacingMergeTree. The order service owns this publish because it is
		// the orchestrator that creates order items (they never exist outside
		// an order flow) and it already holds the full context (merchant,
		// product, category) needed to denormalize the event.
		s.publishStatsEvent(ctx, "stats.ecommerce.order_item.event", strconv.Itoa(int(createdItem.OrderItemID)), events.OrderItemEvent{
			OrderItemID: createdItem.OrderItemID,
			OrderID:     order.OrderID,
			MerchantID:  order.MerchantID,
			ProductID:   createdItem.ProductID,
			CategoryID:  product.CategoryID,
			Quantity:    createdItem.Quantity,
			Price:       createdItem.Price,
			EventTime:   time.Now().UTC().Format(time.RFC3339),
		})
	}

	_, err = s.shippingAddressRepository.Create(ctx, &requests.CreateShippingAddressRequest{
		OrderID:        pointerInt32ToInt(order.OrderID),
		Alamat:         req.ShippingAddress.Alamat,
		Provinsi:       req.ShippingAddress.Provinsi,
		Kota:           req.ShippingAddress.Kota,
		Courier:        req.ShippingAddress.Courier,
		ShippingMethod: req.ShippingAddress.ShippingMethod,
		ShippingCost:   req.ShippingAddress.ShippingCost,
		Negara:         req.ShippingAddress.Negara,
	})
	if err != nil {
		return fail(err)
	}

	totalPrice, err := s.orderItemQueryRepository.CalculateTotalPrice(ctx, int(order.OrderID))
	if err != nil {
		return fail(err)
	}

	res, err := s.orderCommandRepository.Update(ctx, &requests.UpdateOrderRecordRequest{
		OrderID:    int(order.OrderID),
		UserID:     req.UserID,
		TotalPrice: int(*totalPrice) + req.ShippingAddress.ShippingCost,
	})
	if err != nil {
		return fail(err)
	}

	logSuccess("Successfully created order", zap.Int("order.id", int(order.OrderID)))

	// Stats pipeline (F3): publish a stats.ecommerce.order.event so
	// stats-writer materializes this order into ClickHouse for the order
	// revenue statistics. Status mirrors the backfill convention ("created").
	s.publishStatsEvent(ctx, "stats.ecommerce.order.event", strconv.Itoa(int(res.OrderID)), events.OrderEvent{
		OrderID:    res.OrderID,
		UserID:     res.UserID,
		MerchantID: res.MerchantID,
		TotalPrice: res.TotalPrice,
		Status:     "created",
		EventTime:  time.Now().UTC().Format(time.RFC3339),
	})

	return &db.CreateOrderRow{
		OrderID:    res.OrderID,
		UserID:     res.UserID,
		MerchantID: res.MerchantID,
		TotalPrice: res.TotalPrice,
		CreatedAt:  res.CreatedAt,
		UpdatedAt:  res.UpdatedAt,
	}, nil
}

// publishStatsEvent marshals a stats domain event into the standard envelope
// (event_id for idempotency + business payload) and persists it in the outbox
// for the relay to publish to Kafka. It is best-effort: the order flow spans
// multiple services so there is no single database transaction that could carry
// the event atomically; a failure here is logged and must be reconciled
// manually or via backfill, never by rolling back the order.
func (s *orderCommandService) publishStatsEvent(ctx context.Context, topic, key string, payload any) {
	if s.outbox == nil {
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("failed to marshal stats event payload", zap.Error(err), zap.String("topic", topic))
		return
	}
	envelope, err := json.Marshal(events.StatsEnvelope{
		EventID: uuid.NewString(),
		Payload: payloadJSON,
	})
	if err != nil {
		s.logger.Error("failed to marshal stats event envelope", zap.Error(err), zap.String("topic", topic))
		return
	}
	if _, err := s.outbox.Create(ctx, topic, key, envelope); err != nil {
		s.logger.Error("failed to enqueue stats event", zap.Error(err), zap.String("topic", topic))
	}
}

func pointerInt32ToInt(v int32) *int {
	res := int(v)
	return &res
}

func reservationOperationID(prefix string, reservation *db.OrderStockReservation) string {
	return fmt.Sprintf("%s-order-%d-product-%d-reservation-%d-at-%d", prefix, reservation.OrderID, reservation.ProductID, reservation.ReservationID, reservation.UpdatedAt.Time.UnixNano())
}

func (s *orderCommandService) Update(ctx context.Context, req *requests.UpdateOrderRequest) (*db.UpdateOrderRow, error) {
	const method = "Update"
	if req == nil || req.OrderID == nil || *req.OrderID <= 0 {
		return nil, sharedErrors.ErrBadRequest.WithMessage("order id is required")
	}

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("order.id", *req.OrderID),
		attribute.Int("user.id", req.UserID))

	defer func() {
		end(status)
	}()

	existingOrder, err := s.orderQueryRepository.FindByID(ctx, *req.OrderID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*db.UpdateOrderRow](s.logger, err, method, span)
	}

	existingItems, err := s.orderItemQueryRepository.FindOrderItemByOrder(ctx, *req.OrderID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*db.UpdateOrderRow](s.logger, err, method, span)
	}
	itemsByID := make(map[int32]*models.OrderItem, len(existingItems))
	for _, existingItem := range existingItems {
		itemsByID[existingItem.OrderItemID] = existingItem
	}
	reservationRows, err := s.stockReservationRepository.GetByOrder(ctx, *req.OrderID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*db.UpdateOrderRow](s.logger, err, method, span)
	}
	reservationQuantityByProduct := make(map[int]int, len(reservationRows))
	for _, reservation := range reservationRows {
		reservationQuantityByProduct[int(reservation.ProductID)] = int(reservation.Quantity)
	}

	// Each successful item mutation registers a best-effort inverse operation.
	// This keeps stock and child rows consistent when a later step fails.
	compensations := make([]func(), 0, len(req.Items))
	rollback := func() {
		for i := len(compensations) - 1; i >= 0; i-- {
			compensations[i]()
		}
	}
	fail := func(err error) (*db.UpdateOrderRow, error) {
		status = "error"
		rollback()
		return errorhandler.HandleError[*db.UpdateOrderRow](s.logger, err, method, span)
	}

	for itemIndex, item := range req.Items {
		if item.Quantity <= 0 {
			return fail(sharedErrors.ErrBadRequest.WithMessage("order item quantity must be greater than zero"))
		}

		if item.OrderItemID > 0 {
			existingItem, ok := itemsByID[int32(item.OrderItemID)]
			if !ok {
				return fail(sharedErrors.ErrBadRequest.WithMessage("order item does not belong to order"))
			}
			if int(existingItem.ProductID) != item.ProductID {
				return fail(sharedErrors.ErrBadRequest.WithMessage("changing an order item's product is not supported"))
			}

			product, err := s.productQueryRepository.FindByID(ctx, item.ProductID)
			if err != nil {
				return fail(err)
			}
			delta := int(existingItem.Quantity) - item.Quantity
			if delta != 0 {
				operationID := fmt.Sprintf("order-update-%d-item-%d-from-%d-to-%d", *req.OrderID, item.OrderItemID, existingItem.Quantity, item.Quantity)
				if _, err = s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, delta, operationID); err != nil {
					return fail(err)
				}
			}
			if _, err = s.stockReservationRepository.UpdateQuantity(ctx, *req.OrderID, item.ProductID, item.Quantity); err != nil {
				if delta != 0 {
					_, _ = s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, -delta, fmt.Sprintf("order-update-reservation-rollback-%d-item-%d", *req.OrderID, item.OrderItemID))
				}
				return fail(err)
			}
			_, err = s.orderItemCommandRepos.Update(ctx, &requests.UpdateOrderItemRecordRequest{
				OrderItemID: item.OrderItemID,
				OrderID:     *req.OrderID,
				ProductID:   item.ProductID,
				Quantity:    item.Quantity,
				Price:       int(product.Price),
			})
			if err != nil {
				if delta != 0 {
					_, _ = s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, -delta, fmt.Sprintf("order-update-rollback-%d-item-%d", *req.OrderID, item.OrderItemID))
				}
				_, _ = s.stockReservationRepository.UpdateQuantity(ctx, *req.OrderID, item.ProductID, int(existingItem.Quantity))
				return fail(err)
			}

			oldQuantity := int(existingItem.Quantity)
			oldPrice := int(existingItem.Price)
			compensations = append(compensations, func() {
				if _, rollbackErr := s.stockReservationRepository.UpdateQuantity(ctx, *req.OrderID, item.ProductID, oldQuantity); rollbackErr != nil {
					s.logger.Error("failed to compensate stock reservation update", zap.Error(rollbackErr), zap.Int("product_id", item.ProductID))
				}
				if _, rollbackErr := s.orderItemCommandRepos.Update(ctx, &requests.UpdateOrderItemRecordRequest{
					OrderItemID: item.OrderItemID,
					OrderID:     *req.OrderID,
					ProductID:   item.ProductID,
					Quantity:    oldQuantity,
					Price:       oldPrice,
				}); rollbackErr != nil {
					s.logger.Error("failed to compensate order item update", zap.Error(rollbackErr), zap.Int("order_item_id", item.OrderItemID))
				}
				if delta != 0 {
					if _, rollbackErr := s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, -delta, fmt.Sprintf("order-update-compensate-%d-item-%d", *req.OrderID, item.OrderItemID)); rollbackErr != nil {
						s.logger.Error("failed to compensate order stock update", zap.Error(rollbackErr), zap.Int("product_id", item.ProductID))
					}
				}
			})
			continue
		}

		product, err := s.productQueryRepository.FindByID(ctx, item.ProductID)
		if err != nil {
			return fail(err)
		}
		operationID := fmt.Sprintf("order-update-new-%d-index-%d-product-%d-quantity-%d", *req.OrderID, itemIndex, item.ProductID, item.Quantity)
		if _, err = s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, -item.Quantity, operationID); err != nil {
			return fail(err)
		}

		createdItem, err := s.orderItemCommandRepos.Create(ctx, &requests.CreateOrderItemRecordRequest{
			OrderID:   *req.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     int(product.Price),
		})
		if err != nil {
			_, _ = s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, item.Quantity, fmt.Sprintf("order-update-new-rollback-%d-product-%d", *req.OrderID, item.ProductID))
			return fail(err)
		}
		oldReservationQuantity, hadReservation := reservationQuantityByProduct[item.ProductID]
		compensations = append(compensations, func() {
			var rollbackErr error
			if hadReservation {
				_, rollbackErr = s.stockReservationRepository.UpdateQuantity(ctx, *req.OrderID, item.ProductID, oldReservationQuantity)
			} else {
				rollbackErr = s.stockReservationRepository.DeleteByOrderProduct(ctx, *req.OrderID, item.ProductID)
			}
			if rollbackErr != nil {
				s.logger.Error("failed to restore compensated stock reservation", zap.Error(rollbackErr), zap.Int("product_id", item.ProductID))
			}
			if _, rollbackErr := s.orderItemCommandRepos.Trash(ctx, int(createdItem.OrderItemID)); rollbackErr == nil {
				if _, rollbackErr = s.orderItemCommandRepos.DeletePermanent(ctx, int(createdItem.OrderItemID)); rollbackErr != nil {
					s.logger.Error("failed to permanently delete compensated order item", zap.Error(rollbackErr), zap.Int32("order_item_id", createdItem.OrderItemID))
				}
			} else {
				s.logger.Error("failed to trash compensated order item", zap.Error(rollbackErr), zap.Int32("order_item_id", createdItem.OrderItemID))
			}
			if _, rollbackErr := s.productCommandRepository.AdjustProductStock(ctx, item.ProductID, item.Quantity, fmt.Sprintf("order-update-new-compensate-%d-product-%d", *req.OrderID, item.ProductID)); rollbackErr != nil {
				s.logger.Error("failed to compensate new order item stock", zap.Error(rollbackErr), zap.Int("product_id", item.ProductID))
			}
		})
		if _, err = s.stockReservationRepository.Upsert(ctx, *req.OrderID, item.ProductID, item.Quantity); err != nil {
			return fail(err)
		}
	}

	if s.shippingQueryRepository == nil {
		return fail(sharedErrors.ErrBadRequest.WithMessage("shipping query dependency is required"))
	}

	var previousShipping *models.ShippingAddress
	var shippingID *int
	if req.ShippingAddress != nil && req.ShippingAddress.ShippingID != nil {
		shippingID = req.ShippingAddress.ShippingID
		shippingRes, lookupErr := s.shippingQueryRepository.FindByID(ctx, *shippingID)
		if lookupErr != nil {
			return fail(lookupErr)
		}
		if shippingRes == nil {
			return fail(sharedErrors.ErrBadRequest.WithMessage("shipping address is required"))
		}
		previousShipping = shippingRes
	} else {
		shippingRes, lookupErr := s.shippingQueryRepository.FindByOrder(ctx, *req.OrderID)
		if lookupErr != nil {
			return fail(lookupErr)
		}
		if shippingRes == nil {
			return fail(sharedErrors.ErrBadRequest.WithMessage("shipping address is required"))
		}
		id := int(shippingRes.ShippingAddressID)
		shippingID = &id
		previousShipping = shippingRes
	}

	if previousShipping.OrderID != existingOrder.OrderID {
		return fail(sharedErrors.ErrBadRequest.WithMessage("shipping address does not belong to order"))
	}

	shippingCost := int(previousShipping.ShippingCost)
	if req.ShippingAddress != nil {
		shippingUpdate, updateErr := s.shippingAddressRepository.Update(ctx, &requests.UpdateShippingAddressRequest{
			ShippingID:     shippingID,
			OrderID:        pointerInt32ToInt(existingOrder.OrderID),
			Alamat:         req.ShippingAddress.Alamat,
			Provinsi:       req.ShippingAddress.Provinsi,
			Kota:           req.ShippingAddress.Kota,
			Courier:        req.ShippingAddress.Courier,
			ShippingMethod: req.ShippingAddress.ShippingMethod,
			ShippingCost:   req.ShippingAddress.ShippingCost,
			Negara:         req.ShippingAddress.Negara,
		})
		if updateErr != nil {
			return fail(updateErr)
		}
		if shippingUpdate == nil {
			return fail(sharedErrors.ErrBadRequest.WithMessage("shipping address update returned no data"))
		}
		shippingCost = int(shippingUpdate.ShippingCost)

		oldShippingID := *shippingID
		oldOrderID := previousShipping.OrderID
		oldAlamat := previousShipping.Alamat
		oldProvinsi := previousShipping.Provinsi
		oldKota := previousShipping.Kota
		oldCourier := previousShipping.Courier
		oldShippingMethod := previousShipping.ShippingMethod
		oldShippingCost := int(previousShipping.ShippingCost)
		oldNegara := previousShipping.Negara
		compensations = append(compensations, func() {
			if _, rollbackErr := s.shippingAddressRepository.Update(ctx, &requests.UpdateShippingAddressRequest{
				ShippingID:     &oldShippingID,
				OrderID:        pointerInt32ToInt(oldOrderID),
				Alamat:         oldAlamat,
				Provinsi:       oldProvinsi,
				Kota:           oldKota,
				Courier:        oldCourier,
				ShippingMethod: oldShippingMethod,
				ShippingCost:   oldShippingCost,
				Negara:         oldNegara,
			}); rollbackErr != nil {
				s.logger.Error("failed to compensate shipping address update", zap.Error(rollbackErr), zap.Int("shipping_id", oldShippingID))
			}
		})
	}

	totalPrice, err := s.orderItemQueryRepository.CalculateTotalPrice(ctx, *req.OrderID)
	if err != nil {
		return fail(err)
	}

	res, err := s.orderCommandRepository.Update(ctx, &requests.UpdateOrderRecordRequest{
		OrderID:    *req.OrderID,
		UserID:     req.UserID,
		TotalPrice: int(*totalPrice) + shippingCost,
	})

	if err != nil {
		return fail(err)
	}

	s.cache.DeleteOrderCache(ctx, *req.OrderID)

	logSuccess("Successfully updated order", zap.Int("order.id", *req.OrderID))

	return res, nil
}

func (s *orderCommandService) Trash(ctx context.Context, orderID int) (*db.Order, error) {
	const method = "Trash"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("orderID", orderID))

	defer func() {
		end(status)
	}()

	reservations, reservationErr := s.stockReservationRepository.GetByOrder(ctx, orderID)
	if reservationErr != nil {
		status = "error"
		return errorhandler.HandleError[*db.Order](s.logger, reservationErr, method, span)
	}
	released := make([]*db.OrderStockReservation, 0, len(reservations))
	restoreReleasedStock := func() {
		for _, reservation := range released {
			if _, rollbackErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), -int(reservation.Quantity), reservationOperationID("order-trash-rollback", reservation)); rollbackErr != nil {
				s.logger.Error("failed to compensate trashed order stock", zap.Error(rollbackErr), zap.Int32("product_id", reservation.ProductID))
			}
			if _, rollbackErr := s.stockReservationRepository.Reserve(ctx, orderID, int(reservation.ProductID)); rollbackErr != nil {
				s.logger.Error("failed to compensate trashed order reservation", zap.Error(rollbackErr), zap.Int32("product_id", reservation.ProductID))
			}
		}
	}
	for _, reservation := range reservations {
		if reservation.Status != "reserved" {
			continue
		}
		// Claim the reservation before changing stock. The conditional SQL
		// transition prevents two concurrent trash operations from both returning
		// the same quantity to inventory.
		if _, releaseErr := s.stockReservationRepository.Release(ctx, orderID, int(reservation.ProductID)); releaseErr != nil {
			if errors.Is(releaseErr, sql.ErrNoRows) {
				continue
			}
			restoreReleasedStock()
			status = "error"
			return errorhandler.HandleError[*db.Order](s.logger, releaseErr, method, span)
		}
		if _, adjustErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), int(reservation.Quantity), reservationOperationID("order-trash-release", reservation)); adjustErr != nil {
			if _, statusErr := s.stockReservationRepository.Reserve(ctx, orderID, int(reservation.ProductID)); statusErr != nil {
				s.logger.Error("failed to compensate current trashed order reservation", zap.Error(statusErr), zap.Int32("product_id", reservation.ProductID))
			}
			restoreReleasedStock()
			status = "error"
			return errorhandler.HandleError[*db.Order](s.logger, adjustErr, method, span)
		}
		released = append(released, reservation)
	}

	order, err := s.orderCommandRepository.Trash(ctx, orderID)
	if err != nil {
		restoreReleasedStock()
		status = "error"
		return errorhandler.HandleError[*db.Order](s.logger, err, method, span)
	}

	s.cache.DeleteOrderCache(ctx, orderID)

	logSuccess("Successfully trashed order", zap.Int("orderID", orderID))

	return order, nil
}

func (s *orderCommandService) Restore(ctx context.Context, orderID int) (*db.Order, error) {
	const method = "Restore"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("orderID", orderID))

	defer func() {
		end(status)
	}()

	// Claim each released reservation before touching inventory. The conditional
	// SQL transition is the concurrency guard: a second restore sees no released
	// row and cannot decrement stock for the same reservation.
	reservations, reservationErr := s.stockReservationRepository.GetByOrder(ctx, orderID)
	if reservationErr != nil {
		_, _ = s.orderCommandRepository.Trash(ctx, orderID)
		status = "error"
		return errorhandler.HandleError[*db.Order](s.logger, reservationErr, method, span)
	}
	reserved := make([]*db.OrderStockReservation, 0, len(reservations))
	rollbackReservations := func() {
		for _, reservation := range reserved {
			if _, rollbackErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), int(reservation.Quantity), reservationOperationID("order-restore-rollback", reservation)); rollbackErr != nil {
				s.logger.Error("failed to compensate restored stock", zap.Error(rollbackErr), zap.Int32("product_id", reservation.ProductID))
			}
			if _, rollbackErr := s.stockReservationRepository.Release(ctx, orderID, int(reservation.ProductID)); rollbackErr != nil {
				s.logger.Error("failed to compensate restored reservation", zap.Error(rollbackErr), zap.Int32("product_id", reservation.ProductID))
			}
		}
	}
	for _, reservation := range reservations {
		if reservation.Status != "released" {
			continue
		}
		if _, reserveErr := s.stockReservationRepository.Reserve(ctx, orderID, int(reservation.ProductID)); reserveErr != nil {
			if errors.Is(reserveErr, sql.ErrNoRows) {
				continue
			}
			rollbackReservations()
			status = "error"
			return errorhandler.HandleError[*db.Order](s.logger, reserveErr, method, span)
		}
		if _, adjustErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), -int(reservation.Quantity), reservationOperationID("order-restore-reserve", reservation)); adjustErr != nil {
			if _, rollbackErr := s.stockReservationRepository.Release(ctx, orderID, int(reservation.ProductID)); rollbackErr != nil {
				s.logger.Error("failed to compensate current restored reservation", zap.Error(rollbackErr), zap.Int32("product_id", reservation.ProductID))
			}
			rollbackReservations()
			status = "error"
			return errorhandler.HandleError[*db.Order](s.logger, adjustErr, method, span)
		}
		reserved = append(reserved, reservation)
	}

	order, err := s.orderCommandRepository.Restore(ctx, orderID)
	if err != nil {
		rollbackReservations()
		status = "error"
		return errorhandler.HandleError[*db.Order](s.logger, err, method, span)
	}

	s.cache.DeleteOrderCache(ctx, orderID)

	logSuccess("Successfully restored order", zap.Int("orderID", orderID))

	return order, nil
}

func (s *orderCommandService) DeletePermanent(ctx context.Context, orderID int) (bool, error) {
	const method = "DeletePermanent"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("orderID", orderID))

	defer func() {
		end(status)
	}()

	// Permanent deletion is orchestrated across the per-service databases. The
	// local SQL statement deletes the trashed order and its reservation ledger
	// atomically; the child rows (order items, transactions, shipping addresses)
	// live in other services' databases, so they are purged via gRPC to the
	// owning services. No stock is mutated here — Trash already returned
	// inventory.
	success, err := s.orderCommandRepository.DeletePermanentWithChildren(ctx, orderID)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}

	// Purge children owned by other services (order-scoped, so they only touch
	// rows of this trashed order). Failures here surface as delete errors rather
	// than silently orphaning children.
	if _, err = s.orderItemCommandRepos.DeleteByOrderIDPermanent(ctx, orderID); err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}
	if _, err = s.transactionCommandRepos.DeleteByOrderIDPermanent(ctx, orderID); err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}
	if _, err = s.shippingAddressRepository.DeleteByOrderIDPermanent(ctx, orderID); err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}

	s.cache.DeleteOrderCache(ctx, orderID)

	logSuccess("Successfully deleted order permanently", zap.Int("orderID", orderID))

	return success, nil
}

func (s *orderCommandService) RestoreAll(ctx context.Context) (bool, error) {
	const method = "RestoreAll"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	trashedOrders, err := s.orderCommandRepository.FindTrashed(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}
	for _, trashedOrder := range trashedOrders {
		if _, restoreErr := s.Restore(ctx, int(trashedOrder.OrderID)); restoreErr != nil {
			// A concurrent restore has already claimed this order, so continue.
			// Other failures are returned instead of silently reporting success.
			if errors.Is(restoreErr, sql.ErrNoRows) || errors.Is(restoreErr, order_errors.ErrOrderNotFound) {
				continue
			}
			status = "error"
			return errorhandler.HandleError[bool](s.logger, restoreErr, method, span)
		}
	}

	success := true

	s.cache.InvalidateOrderCache(ctx)
	logSuccess("Successfully restored all orders")

	return success, nil
}

// ReconcileResult reports how many reservations durable reconciliation repaired.
type ReconcileResult struct {
	ReReserved int
	Released   int
}

// CleanupResult reports how many rows the retention policy removed.
type CleanupResult struct {
	ReleasedReservationsRemoved int64
	AdjustmentsRemoved          int64
}

// ReconcileStockReservations repairs drift between reservation status and order
// lifecycle that best-effort compensation may have left behind:
//   - a reservation marked released while its order is still active must be
//     re-reserved (and stock reserved again);
//   - a reservation still marked reserved while its order is trashed must be
//     released (and stock returned).
//
// Every repair is idempotent through its operation ID, so a failed run can be
// re-run safely. Failures are logged and the remaining rows are still processed.
func (s *orderCommandService) ReconcileStockReservations(ctx context.Context) (*ReconcileResult, error) {
	const method = "ReconcileStockReservations"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	result := &ReconcileResult{}

	// 1. Released reservations belonging to active orders: re-reserve stock.
	releasedOnActive, err := s.stockReservationRepository.GetReleasedForActiveOrders(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*ReconcileResult](s.logger, err, method, span)
	}
	for _, reservation := range releasedOnActive {
		if _, reserveErr := s.stockReservationRepository.Reserve(ctx, int(reservation.OrderID), int(reservation.ProductID)); reserveErr != nil {
			if !errors.Is(reserveErr, sql.ErrNoRows) {
				s.logger.Error("failed to claim reservation during reconciliation", zap.Error(reserveErr), zap.Int32("order_id", reservation.OrderID), zap.Int32("product_id", reservation.ProductID))
			}
			continue
		}
		if _, adjustErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), -int(reservation.Quantity), reservationOperationID("order-reconcile-reserve", reservation)); adjustErr != nil {
			_, _ = s.stockReservationRepository.Release(ctx, int(reservation.OrderID), int(reservation.ProductID))
			s.logger.Error("failed to re-reserve stock during reconciliation", zap.Error(adjustErr), zap.Int32("order_id", reservation.OrderID), zap.Int32("product_id", reservation.ProductID))
			continue
		}
		result.ReReserved++
	}

	// 2. Reserved reservations belonging to trashed orders: release stock.
	reservedOnTrashed, err := s.stockReservationRepository.GetReservedForTrashedOrders(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*ReconcileResult](s.logger, err, method, span)
	}
	for _, reservation := range reservedOnTrashed {
		if _, releaseErr := s.stockReservationRepository.Release(ctx, int(reservation.OrderID), int(reservation.ProductID)); releaseErr != nil {
			if !errors.Is(releaseErr, sql.ErrNoRows) {
				s.logger.Error("failed to claim release during reconciliation", zap.Error(releaseErr), zap.Int32("order_id", reservation.OrderID), zap.Int32("product_id", reservation.ProductID))
			}
			continue
		}
		if _, adjustErr := s.productCommandRepository.AdjustProductStock(ctx, int(reservation.ProductID), int(reservation.Quantity), reservationOperationID("order-reconcile-release", reservation)); adjustErr != nil {
			_, _ = s.stockReservationRepository.Reserve(ctx, int(reservation.OrderID), int(reservation.ProductID))
			s.logger.Error("failed to release stock during reconciliation", zap.Error(adjustErr), zap.Int32("order_id", reservation.OrderID), zap.Int32("product_id", reservation.ProductID))
			continue
		}
		result.Released++
	}

	logSuccess("Successfully reconciled stock reservations", zap.Int("re_reserved", result.ReReserved), zap.Int("released", result.Released))

	return result, nil
}

// CleanupIdempotencyRecords applies the retention policy to the idempotency
// ledger and to released reservations of trashed orders. Rows older than the
// retention window are purged so the tables stay bounded without touching fresh
// rows that in-flight retries may still reference.
func (s *orderCommandService) CleanupIdempotencyRecords(ctx context.Context, retentionDays int) (*CleanupResult, error) {
	const method = "CleanupIdempotencyRecords"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("retention_days", retentionDays))

	defer func() {
		end(status)
	}()

	if retentionDays <= 0 {
		retentionDays = 7
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	releasedRemoved, err := s.stockReservationRepository.DeleteOldReleasedReservations(ctx, cutoff)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*CleanupResult](s.logger, err, method, span)
	}

	adjustmentsRemoved, err := s.stockReservationRepository.DeleteOldProductStockAdjustments(ctx, cutoff)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[*CleanupResult](s.logger, err, method, span)
	}

	logSuccess("Successfully cleaned up idempotency records", zap.Int64("released_reservations_removed", releasedRemoved), zap.Int64("adjustments_removed", adjustmentsRemoved))

	return &CleanupResult{
		ReleasedReservationsRemoved: releasedRemoved,
		AdjustmentsRemoved:          adjustmentsRemoved,
	}, nil
}

func (s *orderCommandService) DeleteAll(ctx context.Context) (bool, error) {
	const method = "DeleteAll"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method)

	defer func() {
		end(status)
	}()

	// Purge each trashed order through the guarded, order-scoped path. Calling
	// child repositories' DeleteAll methods here would destroy active orders'
	// items, transactions, and shipping addresses because those queries are
	// global, not filtered by trashed order.
	trashedOrders, err := s.orderCommandRepository.FindTrashed(ctx)
	if err != nil {
		status = "error"
		return errorhandler.HandleError[bool](s.logger, err, method, span)
	}
	for _, trashedOrder := range trashedOrders {
		if _, deleteErr := s.DeletePermanent(ctx, int(trashedOrder.OrderID)); deleteErr != nil {
			status = "error"
			return errorhandler.HandleError[bool](s.logger, deleteErr, method, span)
		}
	}

	success := true

	s.cache.InvalidateOrderCache(ctx)
	logSuccess("Successfully deleted all orders permanently")

	return success, nil
}
