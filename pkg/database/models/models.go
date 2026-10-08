// Package models defines the cross-service domain shapes returned by the
// pkg/adapter gRPC adapters. Consuming services depend on these structs (not
// on generated pb types) so the proto surface stays behind the adapter.
package models

import "time"

type User struct {
	UserID    int32      `json:"user_id"`
	Firstname string     `json:"firstname"`
	Lastname  string     `json:"lastname"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type Role struct {
	RoleID    int32      `json:"role_id"`
	RoleName  string     `json:"role_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserRole struct {
	UserRoleID int32 `json:"user_role_id"`
	UserID     int32 `json:"user_id"`
	RoleID     int32 `json:"role_id"`
}

type Merchant struct {
	MerchantID   int32      `json:"merchant_id"`
	UserID       int32      `json:"user_id"`
	Name         string     `json:"name"`
	Description  *string    `json:"description"`
	Address      *string    `json:"address"`
	ContactEmail *string    `json:"contact_email"`
	ContactPhone *string    `json:"contact_phone"`
	Status       string     `json:"status"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type Category struct {
	CategoryID    int32      `json:"category_id"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	SlugCategory  *string    `json:"slug_category"`
	ImageCategory *string    `json:"image_category"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type Product struct {
	ProductID     int32      `json:"product_id"`
	MerchantID    int32      `json:"merchant_id"`
	CategoryID    int32      `json:"category_id"`
	Name          string     `json:"name"`
	Description   *string    `json:"description"`
	Price         int32      `json:"price"`
	CountInStock  int32      `json:"count_in_stock"`
	Weight        *int32     `json:"weight"`
	ImageProduct  *string    `json:"image_product"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

type Order struct {
	OrderID    int32      `json:"order_id"`
	UserID     int32      `json:"user_id"`
	MerchantID int32      `json:"merchant_id"`
	TotalPrice int32      `json:"total_price"`
	CreatedAt  *time.Time `json:"created_at"`
}

type OrderItem struct {
	OrderItemID int32      `json:"order_item_id"`
	OrderID     int32      `json:"order_id"`
	ProductID   int32      `json:"product_id"`
	Quantity    int32      `json:"quantity"`
	Price       int32      `json:"price"`
	CreatedAt   *time.Time `json:"created_at"`
}

// Transaction is the cross-service shape returned by the transaction gRPC
// adapter (used by the stats backfill). Status mirrors the payment_status
// column from the source OLTP table.
type Transaction struct {
	TransactionID int32      `json:"transaction_id"`
	OrderID       int32      `json:"order_id"`
	MerchantID    int32      `json:"merchant_id"`
	PaymentMethod string     `json:"payment_method"`
	Amount        int32      `json:"amount"`
	Status        string     `json:"status"`
	CreatedAt     *time.Time `json:"created_at"`
}

type ShippingAddress struct {
	ShippingAddressID int32   `json:"shipping_address_id"`
	OrderID           int32   `json:"order_id"`
	Alamat            string  `json:"alamat"`
	Provinsi          string  `json:"provinsi"`
	Kota              string  `json:"kota"`
	Negara            string  `json:"negara"`
	Courier           string  `json:"courier"`
	ShippingMethod    string  `json:"shipping_method"`
	ShippingCost      float64 `json:"shipping_cost"`
}
