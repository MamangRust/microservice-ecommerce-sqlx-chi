package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/response"
	"github.com/google/uuid"
	"go.uber.org/zap"

	mencache "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/cache"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
)

type RoleValidator struct {
	kafka         *kafka.Kafka
	logger        logger.LoggerInterface
	requestTopic  string
	responseTopic string
	timeout       time.Duration

	responseChans map[string]chan *response.RoleResponsePayload

	mu sync.RWMutex

	cache mencache.RoleCache
}

func NewRoleValidator(k *kafka.Kafka, requestTopic, responseTopic string, timeout time.Duration, logger logger.LoggerInterface, cache mencache.RoleCache) *RoleValidator {
	v := &RoleValidator{
		kafka:         k,
		requestTopic:  requestTopic,
		responseTopic: responseTopic,
		timeout:       timeout,
		cache:         cache,
		responseChans: make(map[string]chan *response.RoleResponsePayload),
		logger:        logger,
	}

	if k != nil {
		handler := &roleResponseHandler{validator: v}
		go func() {
			err := k.StartConsumers([]string{responseTopic}, "role-validator-gateway", handler)
			if err != nil {
				v.logger.Fatal("Failed to start kafka consumer", zap.Error(err))
				panic("failed to start kafka consumer: " + err.Error())
			}
		}()
	}

	return v
}

func (v *RoleValidator) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDVal := httpx.Get(r, "user_id")
			v.logger.Debug("Validating user role", zap.Any("user_id", userIDVal))
			if userIDVal == nil {
				v.logger.Error("User ID not found in context")
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "User ID not found in context"))
				return
			}
			userID, err := v.extractUserID(userIDVal)
			if err != nil {
				v.logger.Error("Invalid User ID format", zap.Any("value", userIDVal), zap.Error(err))
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "Invalid User ID format"))
				return
			}

			if roles, found := v.cache.GetRoleCache(r.Context(), strconv.Itoa(userID)); found {
				v.logger.Debug("Role found in cache", zap.Int("user_id", userID), zap.Strings("roles", roles))
				r = r.WithContext(httpx.SetValue(r.Context(), "role_names", roles))
				next.ServeHTTP(w, r)
				return
			}

			correlationID := uuid.NewString()
			v.logger.Info("Validating user role via Kafka", zap.Int("user_id", userID), zap.String("correlation_id", correlationID))

			respChan := make(chan *response.RoleResponsePayload, 1)

			v.mu.Lock()
			v.responseChans[correlationID] = respChan
			v.mu.Unlock()

			defer func() {
				v.mu.Lock()
				delete(v.responseChans, correlationID)
				close(respChan)
				v.mu.Unlock()
			}()

			if err := v.sendValidationRequest(userID, correlationID); err != nil {
				httpx.WriteHTTPError(w, err)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), v.timeout)
			defer cancel()
			select {
			case roleResponse := <-respChan:
				if roleResponse == nil {
					v.logger.Error("Received nil response", zap.String("correlation_id", correlationID))
					httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusInternalServerError, "Invalid response received"))
					return
				}
				if !roleResponse.Valid || len(roleResponse.RoleNames) == 0 {
					v.logger.Debug("Role validation failed",
						zap.Int("user_id", userID),
						zap.String("correlation_id", correlationID),
						zap.Bool("valid", roleResponse.Valid),
						zap.Int("role_count", len(roleResponse.RoleNames)))
					httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusUnauthorized, "Role validation failed"))
					return
				}

				v.logger.Info("Role validation success",
					zap.Int("user_id", userID),
					zap.Strings("roles", roleResponse.RoleNames),
					zap.String("correlation_id", correlationID))

				v.cache.SetRoleCache(ctx, strconv.Itoa(userID), roleResponse.RoleNames)

				r = r.WithContext(httpx.SetValue(r.Context(), "role_names", roleResponse.RoleNames))
				next.ServeHTTP(w, r)

			case <-ctx.Done():
				v.logger.Error("Timeout waiting for Kafka response",
					zap.String("correlation_id", correlationID),
					zap.Duration("timeout", v.timeout))
				httpx.WriteHTTPError(w, httpx.NewHTTPError(http.StatusRequestTimeout, "Timeout waiting for role validation"))
			}
		})
	}
}

func (v *RoleValidator) extractUserID(userIDVal interface{}) (int, error) {
	switch val := userIDVal.(type) {
	case float64:
		return int(val), nil
	case int:
		return val, nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, httpx.NewHTTPError(http.StatusUnauthorized, "Unknown user ID type")
	}
}

func (v *RoleValidator) sendValidationRequest(userID int, correlationID string) error {
	payload := requests.RoleRequestPayload{
		UserID:        userID,
		CorrelationID: correlationID,
		ReplyTopic:    v.responseTopic,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		v.logger.Error("Failed to encode payload", zap.Error(err), zap.String("correlation_id", correlationID))
		return httpx.NewHTTPError(http.StatusInternalServerError, "Failed to encode payload")
	}

	err = v.kafka.SendMessage(v.requestTopic, correlationID, data)
	if err != nil {
		v.logger.Error("Failed to send Kafka message", zap.Error(err), zap.String("correlation_id", correlationID))
		return httpx.NewHTTPError(http.StatusInternalServerError, "Failed to send Kafka message")
	}
	v.logger.Info("Kafka message sent for role validation",
		zap.String("topic", v.requestTopic),
		zap.String("correlation_id", correlationID))
	return nil
}
