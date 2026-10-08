package repository

import (
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	ShippingAddressQuery   ShippingAddressQueryRepository
	ShippingAddressCommand ShippingAddressCommandRepository
}

func NewRepositories(DB *sqlx.DB) *Repositories {
	return &Repositories{
		ShippingAddressQuery:   NewShippingAddressQueryRepository(DB),
		ShippingAddressCommand: NewShippingAddressCommandRepository(DB),
	}
}
