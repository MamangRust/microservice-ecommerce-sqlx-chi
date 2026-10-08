package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	BannerQuery   BannerQueryRepository
	BannerCommand BannerCommandRepository
}

func NewRepositories(db *sqlx.DB) *Repositories {

	return &Repositories{
		BannerQuery:   NewBannerQueryRepository(db),
		BannerCommand: NewBannerCommandRepository(db),
	}
}
