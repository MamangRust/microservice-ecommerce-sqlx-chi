-- +goose Up
-- +goose StatementBegin
CREATE TABLE "order_stock_reservations" (
    "reservation_id" SERIAL PRIMARY KEY,
    "order_id" INT NOT NULL REFERENCES "orders" ("order_id"),
    "product_id" INT NOT NULL,
    "quantity" INT NOT NULL CHECK ("quantity" > 0),
    "status" VARCHAR(20) NOT NULL DEFAULT 'reserved' CHECK ("status" IN ('reserved', 'released')),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "uq_order_stock_reservation" UNIQUE ("order_id", "product_id")
);
CREATE INDEX "idx_order_stock_reservations_order_id" ON "order_stock_reservations"("order_id");
CREATE INDEX "idx_order_stock_reservations_status" ON "order_stock_reservations"("status");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "order_stock_reservations";
-- +goose StatementEnd
