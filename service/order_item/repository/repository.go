package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	OrderItemQuery   OrderItemQueryRepository
	OrderItemCommand OrderItemCommandRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		OrderItemQuery:   NewOrderItemQueryRepository(DB),
		OrderItemCommand: NewOrderItemCommandRepository(DB),
	}
}
