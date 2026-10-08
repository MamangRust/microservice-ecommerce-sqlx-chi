-- ClickHouse Schema for Ecommerce Stats (F3)
--
-- The stats pipeline materializes OLTP events into ClickHouse:
--   domain services --(Kafka stats.ecommerce.*.event)--> stats-writer --> ClickHouse
--   apigateway --(gRPC)--> stats-reader --> ClickHouse
--
-- ReplacingMergeTree(event_version) dedupes at-least-once redeliveries: a row
-- with the same ORDER BY key and a newer event_version replaces the older one.
-- backfill writes event_version = unix timestamp of the run so a re-backfill
-- supersedes prior rows instead of duplicating aggregates.

CREATE TABLE IF NOT EXISTS order_events
(
    event_id      UUID,
    order_id      UInt64,
    user_id       UInt64,
    merchant_id   UInt64,
    total_price   Int64,
    created_at    DateTime,
    event_version UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (order_id, event_id);

CREATE TABLE IF NOT EXISTS order_item_events
(
    event_id       UUID,
    order_item_id  UInt64,
    order_id       UInt64,
    merchant_id    UInt64,
    product_id     UInt64,
    category_id    UInt64,
    category_name  String,
    quantity       UInt32,
    price          Int64,
    created_at     DateTime,
    event_version  UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (order_item_id, event_id);

CREATE TABLE IF NOT EXISTS transaction_events
(
    event_id       UUID,
    transaction_id UInt64,
    order_id       UInt64,
    merchant_id    UInt64,
    payment_method String,
    amount         Int64,
    status         String,
    created_at     DateTime,
    event_version  UInt64
) ENGINE = ReplacingMergeTree(event_version)
ORDER BY (transaction_id, event_id);
