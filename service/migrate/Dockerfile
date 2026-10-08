FROM golang:1.25.7-bookworm AS builder

WORKDIR /app

# Copy workspace and dependency manifests
COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY shared/go.mod shared/go.sum* ./shared/
COPY tests/go.mod tests/go.sum* ./tests/

# Copy all service go.mod files (and keep the per-service migration SQL files,
# which the migrate binary globs at runtime from /app/service/<svc>/...).
COPY service/ ./service/
RUN find service -type f \
    ! -name "go.mod" ! -name "go.sum" \
    ! -path "*/database/migration/*" -delete

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /app

# Copy source code
COPY pkg/ ./pkg/
COPY shared/ ./shared/
COPY service/migrate/ ./service/migrate/

WORKDIR /app/service/migrate

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate_bin ./cmd/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 appuser && \
    adduser -D -s /bin/sh -u 1000 -G appuser appuser

WORKDIR /app

RUN mkdir -p /var/log/app && chown -R 1000:1000 /var/log/app

COPY --from=builder /app/migrate_bin ./migrate
# Ship the per-service migration files so the binary can glob them from /app.
COPY --from=builder /app/service ./service/

RUN chown -R 1000:1000 ./migrate ./service

USER appuser

CMD ["./migrate"]
