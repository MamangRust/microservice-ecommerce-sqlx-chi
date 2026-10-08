package usecase

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-stats-writer/repository"
	"github.com/MamangRust/microservice-ecommerce-shared/domain/events"
)

// UseCase bridges the Kafka consumer and the ClickHouse repository. It also
// derives the ReplacingMergeTree version: live events use the business event
// time so a newer status-change supersedes an older one.
type UseCase interface {
	SaveOrderEvent(ctx context.Context, eventID string, event events.OrderEvent) error
	SaveOrderItemEvent(ctx context.Context, eventID string, event events.OrderItemEvent) error
	SaveTransactionEvent(ctx context.Context, eventID string, event events.TransactionEvent) error

	Close() error
}

type statsUseCase struct {
	repo repository.Repository
}

func NewStatsUseCase(repo repository.Repository) UseCase {
	return &statsUseCase{repo: repo}
}

func (u *statsUseCase) SaveOrderEvent(ctx context.Context, eventID string, event events.OrderEvent) error {
	return u.repo.InsertOrderEvent(ctx, eventID, eventVersion(event.EventTime), event)
}

func (u *statsUseCase) SaveOrderItemEvent(ctx context.Context, eventID string, event events.OrderItemEvent) error {
	return u.repo.InsertOrderItemEvent(ctx, eventID, eventVersion(event.EventTime), event)
}

func (u *statsUseCase) SaveTransactionEvent(ctx context.Context, eventID string, event events.TransactionEvent) error {
	return u.repo.InsertTransactionEvent(ctx, eventID, eventVersion(event.EventTime), event)
}

func (u *statsUseCase) Close() error {
	return u.repo.Close()
}

func eventVersion(eventTime string) uint64 {
	if eventTime == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339, eventTime)
	if err != nil {
		return 0
	}
	return uint64(t.Unix())
}
