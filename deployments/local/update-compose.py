#!/usr/bin/env python3
"""Rewrite deployments/local/docker-compose.yml for per-service databases.

Each domain service gets its own PostgreSQL instance (auth-db, role-db, ...).
The service containers override DB_HOST/DB_NAME so every service only talks to
its own database. The single shared `postgres` + global `migrate` services are
removed (migrations now run per service at startup).
"""
import re
import sys

COMPOSE = "deployments/local/docker-compose.yml"

# (compose service block name, db service name, db name, volume name)
DB_SERVICES = [
    ("auth", "auth-db", "auth_db"),
    ("role", "role-db", "role_db"),
    ("user", "user-db", "user_db"),
    ("email", "email-db", "email_db"),
    ("category", "category-db", "category_db"),
    ("merchant", "merchant-db", "merchant_db"),
    ("merchant_award", "merchant_award-db", "merchant_award_db"),
    ("merchant_business", "merchant_business-db", "merchant_business_db"),
    ("merchant_detail", "merchant_detail-db", "merchant_detail_db"),
    ("merchant_policy", "merchant_policy-db", "merchant_policy_db"),
    ("order", "order-db", "order_db"),
    ("order-item", "order_item-db", "order_item_db"),
    ("product", "product-db", "product_db"),
    ("transaction", "transaction-db", "transaction_db"),
    ("cart", "cart-db", "cart_db"),
    ("review", "review-db", "review_db"),
    ("review_detail", "review_detail-db", "review_detail_db"),
    ("slider", "slider-db", "slider_db"),
    ("shipping_address", "shipping_address-db", "shipping_address_db"),
    ("banner", "banner-db", "banner_db"),
]

DB_BLOCK = """  {dbsvc}:
    image: postgres:17-alpine
    container_name: postgres_{svc}
    command: ["postgres", "-c", "max_connections=400", "-c", "shared_buffers=256MB"]
    environment:
      POSTGRES_USER: DRAGON
      POSTGRES_PASSWORD: DRAGON
      POSTGRES_DB: {dbname}
    ports:
      - "{port}:5432"
    volumes:
      - postgres_{svc}_data:/var/lib/postgresql/data
    networks:
      - app_ecommerce_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U DRAGON -d {dbname}"]
      interval: 5s
      timeout: 5s
      retries: 5
"""

TOP_LEVEL = re.compile(r"^  [a-zA-Z0-9_-]+:\s*$")


def split_blocks(lines):
    """Split top-level service blocks. Returns list of (start, end, name|None)."""
    blocks = []
    i = 0
    while i < len(lines):
        line = lines[i].rstrip("\n")
        if TOP_LEVEL.match(line):
            start = i
            name = line.strip().rstrip(":").strip()
            i += 1
            while i < len(lines) and not TOP_LEVEL.match(lines[i].rstrip("\n")):
                i += 1
            blocks.append((start, i, name))
        else:
            i += 1
    return blocks


def find_block(lines, name):
    for start, end, n in split_blocks(lines):
        if n == name:
            return start, end
    return None


def main():
    with open(COMPOSE) as f:
        lines = f.readlines()

    # 1. Remove the shared `postgres` and `migrate` blocks.
    for name in ("postgres", "migrate"):
        rng = find_block(lines, name)
        if rng is None:
            print(f"WARN: block {name} not found", file=sys.stderr)
            continue
        start, end = rng
        del lines[start:end]

    # 2. Insert per-service db blocks right after the `services:` line.
    services_idx = next(i for i, l in enumerate(lines) if l.startswith("services:"))
    db_port = 5543
    db_blocks = []
    for svc, dbsvc, dbname in DB_SERVICES:
        db_blocks.append(DB_BLOCK.format(svc=svc, dbsvc=dbsvc, dbname=dbname, port=db_port))
        db_port += 1
    lines[services_idx + 1 : services_idx + 1] = "".join(db_blocks).splitlines(keepends=True)

    # 3. For each service container: add DB_HOST/DB_NAME override + point
    #    depends_on at its own db.
    text = "".join(lines)
    for svc, dbsvc, dbname in DB_SERVICES:
        m = re.search(r"(^  %s:\n(?:.*\n)*?)(?=^  [a-zA-Z0-9_-]+:|\Z)" % re.escape(svc), text, re.M)
        if not m:
            print(f"WARN: service block {svc} not found", file=sys.stderr)
            continue
        block = m.group(1)
        # add DB_HOST/DB_NAME to environment (after the first `environment:` line)
        if "DB_HOST=" not in block:
            block = re.sub(
                r"(^    environment:\n)",
                r"\1      - DB_HOST=%s\n      - DB_NAME=%s\n" % (dbsvc, dbname),
                block,
                count=1,
                flags=re.M,
            )
        # replace `postgres:` depends_on entry with own db
        block = re.sub(
            r"^      postgres:\n        condition: service_healthy\n",
            "      %s:\n        condition: service_healthy\n" % dbsvc,
            block,
            flags=re.M,
        )
        text = text[: m.start()] + block + text[m.end() :]

    # 4. apigateway: drop the leftover `postgres:` depends_on (gateway has no db).
    text = re.sub(
        r"^      postgres:\n        condition: service_healthy\n", "", text, count=1, flags=re.M
    )

    # 5. Add postgres volumes to the volumes: section.
    volume_names = ["postgres_%s_data" % svc for svc, _, _ in DB_SERVICES]
    m = re.search(r"^(volumes:\n)", text, re.M)
    if m:
        existing = [v.strip() for v in text[m.end():].splitlines() if v.strip()]
        new_vols = [v for v in volume_names if v not in existing]
        insert = "".join("  %s:\n" % v for v in new_vols)
        text = text[: m.end()] + insert + text[m.end():]

    with open(COMPOSE, "w") as f:
        f.write(text)

    print("compose updated OK")


if __name__ == "__main__":
    main()
