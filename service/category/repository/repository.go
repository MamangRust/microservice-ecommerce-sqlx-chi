package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	CategoryQuery   CategoryQueryRepository
	CategoryCommand CategoryCommandRepository
	// F5: legacy OLTP stats repositories removed.
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		CategoryQuery:   NewCategoryQueryRepository(DB),
		CategoryCommand: NewCategoryCommandRepository(DB),
	}
}
