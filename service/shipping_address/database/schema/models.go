// Hand-written model definitions migrated from sqlc-generated sources.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type ShippingAddress struct {
	ShippingAddressID int32            `db:"shipping_address_id" json:"shipping_address_id"`
	OrderID           int32            `db:"order_id" json:"order_id"`
	Alamat            string           `db:"alamat" json:"alamat"`
	Provinsi          string           `db:"provinsi" json:"provinsi"`
	Negara            string           `db:"negara" json:"negara"`
	Kota              string           `db:"kota" json:"kota"`
	Courier           string           `db:"courier" json:"courier"`
	ShippingMethod    string           `db:"shipping_method" json:"shipping_method"`
	ShippingCost      float64          `db:"shipping_cost" json:"shipping_cost"`
	CreatedAt         pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt         pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt         pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
