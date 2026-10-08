package service

import (
	"context"
	"strconv"
	"time"

	mencache "github.com/MamangRust/microservice-ecommerce-auth/cache"
	models "github.com/MamangRust/microservice-ecommerce-pkg/database/models"
	"github.com/MamangRust/microservice-ecommerce-auth/repository"

	"github.com/MamangRust/microservice-ecommerce-pkg/email"
	"github.com/MamangRust/microservice-ecommerce-pkg/event"
	"github.com/MamangRust/microservice-ecommerce-pkg/hash"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/MamangRust/microservice-ecommerce-pkg/outbox"
	randomstring "github.com/MamangRust/microservice-ecommerce-pkg/randomstring"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-ecommerce-shared/errorhandler"
	user_errors "github.com/MamangRust/microservice-ecommerce-shared/errors/user_errors"

	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type RegisterServiceDeps struct {
	Cache mencache.RegisterCache

	User repository.UserRepository

	Hash hash.HashPassword

	Kafka *kafka.Kafka

	Pool *sqlx.DB

	Outbox *outbox.OutboxService

	Logger logger.LoggerInterface

	Observability observability.TraceLoggerObservability
}

type registerService struct {
	mencache mencache.RegisterCache

	user repository.UserRepository

	hash hash.HashPassword

	kafka *kafka.Kafka

	pool *sqlx.DB

	outbox *outbox.OutboxService

	logger logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

func NewRegisterService(params *RegisterServiceDeps) *registerService {

	return &registerService{
		mencache:      params.Cache,
		user:          params.User,
		hash:          params.Hash,
		kafka:         params.Kafka,
		pool:          params.Pool,
		outbox:        params.Outbox,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *registerService) Register(ctx context.Context, request *requests.RegisterRequest) (*models.User, error) {
	const method = "Register"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("email", request.Email))

	defer func() {
		end(status)
	}()

	existingUser, err := s.user.FindByEmail(ctx, request.Email)
	if err == nil && existingUser != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](
			s.logger,
			user_errors.ErrUserEmailAlready,
			method,
			span,
			zap.String("email", request.Email),
		)
	}

	random, err := randomstring.GenerateRandomString(10)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span)
	}
	request.VerifiedCode = random
	request.IsVerified = false

	newUser, err := s.user.CreateUser(ctx, request)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span)
	}

	htmlBody := email.GenerateEmailHTML(map[string]string{
		"Title":   "Welcome to SanEdge",
		"Message": "Your account has been successfully created.",
		"Button":  "Verify Now",
		"Link":    "https://sanedge.example.com/login?verify_code=" + request.VerifiedCode,
	})

	payloadBytes, err := event.MarshalEmail("auth.register", request.Email, "Welcome to SanEdge", htmlBody)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*models.User](s.logger, err, method, span)
	}

	// Phase 6 — transactional outbox (best-effort enqueue). The user write lives
	// in the user service (gRPC), so the outbox insert cannot share that
	// transaction; it is enqueued here right after the remote writes succeed and
	// the relay guarantees delivery. Direct Kafka remains the fallback when no
	// DB pool is configured (tests/local).
	if s.outbox != nil {
		if enqueueErr := s.outbox.Enqueue(ctx, "email-service-topic-auth-register", strconv.Itoa(int(newUser.UserID)), payloadBytes); enqueueErr != nil {
			s.logger.Error("failed to enqueue registration email to outbox", zap.Error(enqueueErr), zap.String("email", request.Email))
		}
	} else if s.kafka != nil {
		go func() {
			if sendErr := s.kafka.SendMessage("email-service-topic-auth-register", strconv.Itoa(int(newUser.UserID)), payloadBytes); sendErr != nil {
				s.logger.Error("failed to send registration email via kafka", zap.Error(sendErr), zap.String("email", request.Email))
			}
		}()
	}

	s.mencache.SetVerificationCodeCache(ctx, request.Email, random, 15*time.Minute)

	logSuccess("User registered successfully",
		zap.String("email", request.Email),
		zap.String("first_name", request.FirstName),
		zap.String("last_name", request.LastName),
	)

	return newUser, nil
}
