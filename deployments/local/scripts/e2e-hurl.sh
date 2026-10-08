#!/usr/bin/env bash
#
# e2e-hurl.sh — run every E2E hurl suite in tests/hurl/ against the gateway.
#
# Prerequisites:
#   - infra stack running:  docker compose -f deployments/local/docker-compose.yml up -d <svc>-db ... (per-service postgres)
#   - Go services running locally (via `just services-local-start` or run-e2e-mega.sh)
#   - PostgreSQL has seeded data (run `just seeder-local` at least once) so
#     rules_strict.hurl can pick a seeded role-less user.
#
# Usage:
#   bash deployments/local/scripts/e2e-hurl.sh [base_url] [reset_db]
#
# Exits non-zero if any suite fails.
#
# The stack is 1 DB per service: each service owns its own PostgreSQL instance
# (service/<name>/migrations + DB_<CLUSTER>_* env keys). The reset below drops
# the schema in every service DB, re-runs each service's own migrations, then
# runs the per-service seeder (idempotent, no artificial delays).

set -uo pipefail

BASE_URL="${1:-http://localhost:5000}"
RESET_DB="${2:-yes}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
HURL_DIR="$PROJECT_ROOT/tests/hurl"
TEST_IMAGE="$HURL_DIR/assets/test.png"

TS=$(date +%s)
TEST_PASSWORD="HurlPass123"

PASS=0
FAIL=0
FAILED=()

echo "== e2e-hurl: $BASE_URL =="
echo "hurl: $(hurl --version 2>/dev/null | head -1)"

# ---------------------------------------------------------------------------
# Per-service databases: service name -> postgres container / db name / host port
# (ports match deployments/local/docker-compose.yml *-db services).
# ---------------------------------------------------------------------------
declare -A DB_CONTAINER=( [auth]=postgres_auth [role]=postgres_role [user]=postgres_user [email]=postgres_email [category]=postgres_category [merchant]=postgres_merchant [merchant_award]=postgres_merchant_award [merchant_business]=postgres_merchant_business [merchant_detail]=postgres_merchant_detail [merchant_policy]=postgres_merchant_policy [order]=postgres_order [order_item]=postgres_order-item [product]=postgres_product [transaction]=postgres_transaction [cart]=postgres_cart [review]=postgres_review [review_detail]=postgres_review_detail [slider]=postgres_slider [shipping_address]=postgres_shipping_address [banner]=postgres_banner )
declare -A DB_NAME=( [auth]=auth_db [role]=role_db [user]=user_db [email]=email_db [category]=category_db [merchant]=merchant_db [merchant_award]=merchant_award_db [merchant_business]=merchant_business_db [merchant_detail]=merchant_detail_db [merchant_policy]=merchant_policy_db [order]=order_db [order_item]=order_item_db [product]=product_db [transaction]=transaction_db [cart]=cart_db [review]=review_db [review_detail]=review_detail_db [slider]=slider_db [shipping_address]=shipping_address_db [banner]=banner_db )
declare -A DB_PORT=( [auth]=5543 [role]=5544 [user]=5545 [email]=5546 [category]=5547 [merchant]=5548 [merchant_award]=5549 [merchant_business]=5550 [merchant_detail]=5551 [merchant_policy]=5552 [order]=5553 [order_item]=5554 [product]=5555 [transaction]=5556 [cart]=5557 [review]=5558 [review_detail]=5559 [slider]=5560 [shipping_address]=5561 [banner]=5562 )

# Services that own migrations + a database.
ALL_SERVICES=(auth role user email category merchant merchant_award merchant_business merchant_detail merchant_policy order order_item product transaction cart review review_detail slider shipping_address banner)

# ---------------------------------------------------------------------------
# Reset every service database so each run starts from a deterministic state.
# The all-endpoints sweep permanently deletes/trashes seeded rows, so repeated
# runs would otherwise 5xx on already-deleted records.
# ---------------------------------------------------------------------------
if [ "$RESET_DB" = "yes" ]; then
  echo "Resetting databases (drop schema + migrate up per service + seed)..."
  for svc in "${ALL_SERVICES[@]}"; do
    if ! docker exec "${DB_CONTAINER[$svc]}" psql -U DRAGON -d "${DB_NAME[$svc]}" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;" > /tmp/hurl_db_drop_$svc.log 2>&1; then
      echo "WARN: drop schema failed for $svc (see /tmp/hurl_db_drop_$svc.log); continuing"
    fi
  done

  for svc in "${ALL_SERVICES[@]}"; do
    if ! (cd "$PROJECT_ROOT" && DB_HOST=localhost DB_PORT="${DB_PORT[$svc]}" DB_NAME="${DB_NAME[$svc]}" go run service/migrate/cmd/main.go -dir "service/$svc/database/migration" up > /tmp/hurl_migrate_$svc.log 2>&1); then
      echo "ERROR: migrate up failed for $svc (see /tmp/hurl_migrate_$svc.log)"
      exit 1
    fi
  done

  # register auto-assigns ROLE_ADMIN; the seeder only creates Cashier/Manager/Admin/Supplier.
  docker exec postgres_role psql -U DRAGON -d role_db -c \
    "INSERT INTO roles (role_name) VALUES ('ROLE_ADMIN'), ('ROLE_USER') ON CONFLICT DO NOTHING;" >/dev/null 2>&1 || true

  # Drop cached role/auth entries so a fresh run starts from a clean cache.
  docker exec redis-ecommerce-e2e redis-cli -a dragon_knight FLUSHALL >/dev/null 2>&1 || true

  (cd "$PROJECT_ROOT" && go run service/seeder/cmd/main.go > /tmp/hurl_seeder.log 2>&1) || \
    echo "WARN: seeder reported errors (see /tmp/hurl_seeder.log); continuing"

  # F3 stats pipeline: materialize the seeded OLTP orders/order_items/
  # transactions into ClickHouse so stats_strict.hurl (served by stats-reader)
  # sees the data. stats-writer applies the schema at startup, so no separate
  # init step is needed.
  (cd "$PROJECT_ROOT" && go run service/stats_writer/cmd/main.go backfill > /tmp/hurl_backfill.log 2>&1) || \
    echo "WARN: stats backfill reported errors (see /tmp/hurl_backfill.log); continuing"
fi

# ---------------------------------------------------------------------------
# Discover a seeded role-less user for rules_strict.hurl (firstname User1,
# seeded with password "password1"). Register assigns ROLE_ADMIN to every new
# user, so a role-less seeded user is required to prove the 403 path.
# users live in the user service DB (user_db).
# ---------------------------------------------------------------------------
USER_EMAIL=""
if docker exec postgres_user psql -U DRAGON -d user_db -t -A -c \
  "SELECT email FROM users WHERE firstname='User1' AND email LIKE 'user_%' ORDER BY user_id LIMIT 1;" \
  > /tmp/hurl_seed_user.txt 2>/dev/null; then
  USER_EMAIL=$(tr -d ' \r' < /tmp/hurl_seed_user.txt)
  # Guarantee the chosen user is role-less so the 401 denial path is deterministic
  # (register auto-assigns ROLE_ADMIN, and the role seeder may assign random roles).
  # users live in user_db; user_roles live in role_db — resolve the id in the
  # user DB first (the old single-DB subquery would fail against role_db).
  USER_ID=$(docker exec postgres_user psql -U DRAGON -d user_db -t -A -c \
    "SELECT user_id FROM users WHERE email = '$USER_EMAIL' LIMIT 1;" | tr -d ' \r')
  if [ -n "$USER_ID" ]; then
    docker exec postgres_role psql -U DRAGON -d role_db -c \
      "DELETE FROM user_roles WHERE user_id = $USER_ID;" >/dev/null 2>&1 || true
  fi
fi

# ---------------------------------------------------------------------------
# Run every suite. Each file gets its own unique testEmail so register is
# idempotent across re-runs.
# ---------------------------------------------------------------------------
for f in "$HURL_DIR"/*.hurl; do
  name=$(basename "$f")

  testEmail="hurl.${name%.hurl}.$TS@example.com"

  common_vars="--test --jobs 1 --variable baseUrl=$BASE_URL --variable testEmail=$testEmail --variable testPassword=$TEST_PASSWORD --variable testImage=$TEST_IMAGE"

  if [ "$name" = "rules_strict.hurl" ]; then
    if [ -z "$USER_EMAIL" ]; then
      echo "SKIP $name (no seeded role-less user found; run seeder first)"
      continue
    fi
    extra_vars="--variable adminEmail=admin.$TS@example.com --variable adminPassword=$TEST_PASSWORD --variable userEmail=$USER_EMAIL --variable userPassword=password1"
    hurl $common_vars $extra_vars "$f" > "/tmp/hurl_$name.log" 2>&1
  else
    hurl $common_vars "$f" > "/tmp/hurl_$name.log" 2>&1
  fi

  if [ $? -eq 0 ]; then
    PASS=$((PASS + 1))
    echo "PASS $name"
  else
    FAIL=$((FAIL + 1))
    FAILED+=("$name")
    echo "FAIL $name (see /tmp/hurl_$name.log)"
  fi

  # Pace the auth endpoints: the gateway rate-limits register/login with a
  # 10 req/s token bucket shared across all suites, so pause between files.
  sleep 2
done

# ---------------------------------------------------------------------------
# F6: stats flush pause. stats-writer buffers events and flushes ClickHouse in
# batches every ~5s (defaultFlushInterval), so validating stats rows immediately
# after the suites finish would miss the events created by the very last hurl
# files. Wait >= 1 flush interval before reading ClickHouse.
# ---------------------------------------------------------------------------
STATS_FLUSH_INTERVAL_SECS=5
# The compose stack names the container clickhouse-ecommerce; older setups may
# run a shared clickhouse container. Detect whichever is reachable.
CH_CONTAINER=""
if command -v docker >/dev/null 2>&1; then
  for candidate in clickhouse-ecommerce clickhouse; do
    if docker exec "$candidate" clickhouse-client --query "SELECT 1" >/dev/null 2>&1; then
      CH_CONTAINER="$candidate"
      break
    fi
  done
fi
if [ -n "$CH_CONTAINER" ]; then
  echo "Waiting ${STATS_FLUSH_INTERVAL_SECS}s for stats-writer flush before validating stats..."
  sleep $((STATS_FLUSH_INTERVAL_SECS + 2))

  for table in order_events order_item_events transaction_events; do
    row_count=$(docker exec "$CH_CONTAINER" clickhouse-client --database ecommerce --query "SELECT count() FROM $table FINAL" 2>/dev/null | tr -d ' \r')
    echo "stats $table rows: ${row_count:-unknown}"
  done
else
  echo "WARN: no reachable ClickHouse container; skipping stats flush validation"
fi

echo
echo "== RESULT =="
echo "PASS: $PASS   FAIL: $FAIL   TOTAL: $((PASS + FAIL))"

if [ "$FAIL" -gt 0 ]; then
  echo
  echo "Failed suites:"
  for f in "${FAILED[@]}"; do echo "  - $f"; done
  exit 1
fi

echo "All E2E hurl suites passed."
exit 0
