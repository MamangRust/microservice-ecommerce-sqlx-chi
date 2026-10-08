package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	ReviewDetailQuery   ReviewDetailQueryRepository
	ReviewDetailCommand ReviewDetailCommandRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		ReviewDetailQuery:   NewReviewDetailQueryRepository(DB),
		ReviewDetailCommand: NewReviewDetailCommandRepository(DB),
	}
}
