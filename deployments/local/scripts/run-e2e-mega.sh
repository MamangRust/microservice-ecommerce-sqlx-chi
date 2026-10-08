#!/usr/bin/env bash
# Launch all ecommerce Go services locally, wait for health, run the full
# hurl + endpoint-test suites, then clean up. Run from repo root.
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$ROOT"

set -a
# shellcheck disable=SC1091
source "$ROOT/.env"
set +a
# Services are launched with cwd = their own service directory (so their
# ./database/migration resolve), where no .env file exists. APP_ENV=test makes
# dotenv.Viper() skip the config file entirely and rely on the env vars we
# exported above — same pattern as the payment-gateway e2e runner.
export APP_ENV=test

# The ecommerce stack needs 33 Redis DBs (REDIS_DB_* up to 19). The
# payment-gateway redis-local only has 16, so dev/e2e use a dedicated redis on
# port 6380 — see .env (REDIS_PORT=6380).

LOG_DIR="$ROOT/deployments/local/logs"
mkdir -p "$LOG_DIR"

SERVICES=(
  auth role user category merchant merchant_award merchant_business merchant_detail merchant_policy
  order order_item product transaction cart review review_detail slider shipping_address banner
  stats_writer stats_reader email apigateway
)

# Cleanup previous run
pkill -f "$ROOT/bin/" 2>/dev/null || true
sleep 1

# Services run their own migrations at startup from ./database/migration (the same
# layout as the Docker image: WORKDIR /app/service/<name> + COPY migrations),
# so each binary must be launched with cwd = its own service directory.
pids=()
for svc in "${SERVICES[@]}"; do
  (cd "$ROOT/service/$svc" && nohup "$ROOT/bin/$svc" > "$LOG_DIR/$svc.log" 2>&1) &
  pids+=("$!")
done
echo "launched ${#SERVICES[@]} services"

# Wait for health: apigateway /api/auth/hello + all gRPC ports 50051-50070
# (19 domain services + stats-reader).
GW_OK=0
for i in $(seq 1 60); do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 2 http://localhost:5000/api/auth/hello 2>/dev/null)
  ports=$(ss -tln 2>/dev/null | grep -cE ':5005[1-9]|:5006[0-9]|:50070')
  if [ "$code" = "200" ] && [ "$ports" -ge 20 ]; then GW_OK=1; break; fi
  sleep 3
done
echo "apigateway=$code gRPC_ports=$ports"
if [ "$GW_OK" != "1" ]; then
  echo "!!! not healthy; logs with errors:"
  grep -l "panic\|fatal\|address already in use" "$LOG_DIR"/*.log 2>/dev/null | head
  for f in "$LOG_DIR"/*.log; do
    if grep -q "panic\|fatal" "$f" 2>/dev/null; then echo "-- $f"; tail -5 "$f"; fi
  done
  pkill -f "$ROOT/bin/" 2>/dev/null || true
  exit 1
fi

echo "===== e2e-hurl (15 suites) ====="
bash deployments/local/scripts/e2e-hurl.sh http://localhost:5000 > /tmp/e2e-ecommerce-hurl.log 2>&1
HURL_RC=$?
tail -15 /tmp/e2e-ecommerce-hurl.log
echo "e2e-hurl exit=$HURL_RC"

echo "===== endpoint-test (swagger routes) ====="
bash deployments/local/scripts/endpoint-test.sh http://localhost:5000 > /tmp/e2e-ecommerce-endpoint.log 2>&1
EPT_RC=$?
tail -12 /tmp/e2e-ecommerce-endpoint.log
echo "endpoint-test exit=$EPT_RC"

pkill -f "$ROOT/bin/" 2>/dev/null || true
echo "=== cleaned up ==="
exit $(( HURL_RC || EPT_RC ))
