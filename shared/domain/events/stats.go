// Package events defines the domain event payloads used by the stats pipeline
// (F3). Producers publish a StatsEnvelope on a stats.ecommerce.<domain>.event
// topic; stats-writer unwraps it, dedups on EventID, and materializes the
// payload into ClickHouse.
package events

import "encoding/json"

// StatsEnvelope is the standard envelope for stats events. EventID is the
// idempotency key: stats-writer skips events whose EventID was already
// consumed, and ClickHouse keeps it as part of the row's primary key so a
// redelivery cannot duplicate aggregates.
type StatsEnvelope struct {
	EventID string          `json:"event_id"`
	Payload json.RawMessage `json:"payload"`
}

// OrderEvent is published when an order is created or its amount/status
// changes. It feeds the order revenue statistics.
type OrderEvent struct {
	OrderID    int32  `json:"order_id"`
	UserID     int32  `json:"user_id"`
	MerchantID int32  `json:"merchant_id"`
	TotalPrice int32  `json:"total_price"`
	Status     string `json:"status"`
	EventTime  string `json:"event_time"`
}

// OrderItemEvent is published when an order item is created. CategoryID and
// CategoryName are denormalized here so the category pricing statistics can be
// served purely from ClickHouse without a PostgreSQL join.
type OrderItemEvent struct {
	OrderItemID  int32  `json:"order_item_id"`
	OrderID      int32  `json:"order_id"`
	MerchantID   int32  `json:"merchant_id"`
	ProductID    int32  `json:"product_id"`
	CategoryID   int32  `json:"category_id"`
	CategoryName string `json:"category_name"`
	Quantity     int32  `json:"quantity"`
	Price        int32  `json:"price"`
	EventTime    string `json:"event_time"`
}

// TransactionEvent is published when a transaction is created or its status
// changes. It feeds the transaction success/failed and payment-method
// statistics.
type TransactionEvent struct {
	TransactionID int32  `json:"transaction_id"`
	OrderID       int32  `json:"order_id"`
	MerchantID    int32  `json:"merchant_id"`
	PaymentMethod string `json:"payment_method"`
	Amount        int32  `json:"amount"`
	Status        string `json:"status"`
	EventTime     string `json:"event_time"`
}
