package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	SliderQuery   SliderQueryRepository
	SliderCommand SliderCommandRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		SliderQuery:   NewSliderQueryRepository(DB),
		SliderCommand: NewSliderCommandRepository(DB),
	}
}
