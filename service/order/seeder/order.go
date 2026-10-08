package seeder

import (
	"context"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const seedGetOrders = `SELECT
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at,
    COUNT(*) OVER () AS total_count
FROM orders
WHERE
    deleted_at IS NULL
    AND (
        $1::TEXT IS NULL
        OR order_id::TEXT ILIKE '%' || $1 || '%'
        OR total_price::TEXT ILIKE '%' || $1 || '%'
    )
ORDER BY created_at DESC
LIMIT $2
OFFSET
    $3`

const seedCreateOrder = `INSERT INTO
    orders (
        merchant_id,
        user_id,
        total_price
    )
VALUES ($1, $2, $3)
RETURNING
    order_id,
    user_id,
    merchant_id,
    total_price,
    created_at,
    updated_at`

const seedCreateOrderItem = `INSERT INTO
    order_items (
        order_id,
        product_id,
        quantity,
        price
    )
VALUES ($1, $2, $3, $4)
RETURNING
    order_item_id,
    order_id,
    product_id,
    quantity,
    price,
    created_at,
    updated_at`

type seedOrderRow struct {
	OrderID    int32            `db:"order_id"`
	UserID     int32            `db:"user_id"`
	MerchantID int32            `db:"merchant_id"`
	TotalPrice int32            `db:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at"`
	TotalCount int64            `db:"total_count"`
}

type seedOrderItemRow struct {
	OrderItemID int32            `db:"order_item_id"`
	OrderID     int32            `db:"order_id"`
	ProductID   int32            `db:"product_id"`
	Quantity    int32            `db:"quantity"`
	Price       int32            `db:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at"`
}

// orderSeeder seeds orders (order service DB) and their order_items (order_item
// service DB), so it needs both connections.
type orderSeeder struct {
	orderDB     *sqlx.DB
	orderItemDB *sqlx.DB
	ctx         context.Context
	logger      logger.LoggerInterface
}

func NewOrderSeeder(orderDB *sqlx.DB, orderItemDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *orderSeeder {
	return &orderSeeder{
		orderDB:     orderDB,
		orderItemDB: orderItemDB,
		ctx:         ctx,
		logger:      logger,
	}
}

func (r *orderSeeder) Seed() error {
	// Idempotency: skip when orders already exist.
	var existing []seedOrderRow
	err := r.orderDB.SelectContext(r.ctx, &existing, seedGetOrders, "", int32(1), int32(0))
	if err == nil && len(existing) > 0 {
		r.logger.Debug("orders already seeded, skipping")
		return nil
	}

	for i := 1; i <= 8; i++ {
		var order seedOrderRow
		if err := r.orderDB.GetContext(r.ctx, &order, seedCreateOrder,
			int32(i),
			int32(i),
			int32(10000*i),
		); err != nil {
			r.logger.Error("failed to create order", zap.Error(err))
			return err
		}

		var item seedOrderItemRow
		if err := r.orderItemDB.GetContext(r.ctx, &item, seedCreateOrderItem,
			order.OrderID,
			int32(i),
			int32(i),
			int32(10000),
		); err != nil {
			r.logger.Error("failed to create order item", zap.Error(err))
			return err
		}
	}

	r.logger.Info("order & order-item successfully seeded")

	return nil
}
