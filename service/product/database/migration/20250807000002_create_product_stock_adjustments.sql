-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "product_stock_adjustments" (
    "operation_id" VARCHAR(255) PRIMARY KEY,
    "product_id" INT NOT NULL,
    "delta" INT NOT NULL CHECK ("delta" <> 0),
    "created_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS "idx_product_stock_adjustments_product_id" ON "product_stock_adjustments"("product_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS "idx_product_stock_adjustments_product_id";
DROP TABLE IF EXISTS "product_stock_adjustments";
-- +goose StatementEnd
