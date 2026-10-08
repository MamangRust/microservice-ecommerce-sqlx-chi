# Migrasi Apigateway: Echo → Chi

> **Lingkup:** `microservice-ecommerce-grpc/service/apigateway` (+ harness `tests/` dan `pkg/upload_image`)
> **Router:** `github.com/labstack/echo/v4` → `github.com/go-chi/chi/v5`
> **Status:** ✅ Selesai — build ✅ · vet ✅ · unit test ✅
> **Pola:** sama dengan migrasi apigateway `monolith-ecommerce-grpc` (lihat `modular-monolith-grpc-sqlc-restapi/monolith-ecommerce-grpc/service/apigateway/MIGRATION_ECHO_TO_CHI.md`) dan `microservice-pointofsale-grpc` di workspace ini.

---

## 1. Ringkasan

| Item | Jumlah |
|---|---|
| File yang import echo (sebelum) | 76 |
| Paket handler domain yang dimigrasi | 21 (`auth`, `banner`, `cart`, `category`, `merchant`, `merchant_award`, `merchant_business`, `merchant_detail`, `merchant_document`, `merchant_policy`, `merchant_social_link`, `order`, `order_item`, `product`, `review`, `review_detail`, `role`, `shipping_address`, `slider`, `transaction`, `user`) — struktur subpackage (`handler.go`/`query.go`/`command.go`), 63 file |
| Rute terdaftar (terkonversi) | 214 (83 GET · 111 POST · 20 DELETE) — terverifikasi identik dengan baseline |
| Middleware global | 10 (`Trace → HTTPMetrics → Recoverer → RequestID → Logger(JSON) → Pyroscope → CORS → Compress → Secure → RateLimit(/api/auth/) → JWTAuth`) |
| Test harness `tests/` yang dikonversi | 37 file (19 `handler_api_test.go` + 17 `negative_test.go` + `mock_upload.go`) |
| Rute terdampak duplikasi mount | 1 (`/api/order-item`) — rute query didaftarkan langsung di parent router dengan path penuh, command tetap `Route()` (pola monolith) |

## 2. Karakteristik Khusus Ecommerce

1. **Error handler global yang lebih pintar** — `middlewares/errorhandler.go` (`RegisterErrorHandler` yang memetakan `*AppError` shared → status asli) dihapus; logikanya pindah ke adapter `httpx.Handler` (rute unwrapped) dan `apierror.HandleApiError` (rute wrapped). `middlewares/responsewriter.go` ditambahkan untuk metrics/trace (`wrapResponseWriter` pengganti `c.Response().Status`).
2. **JWT menyimpan `user_id` sebagai `int`** — `JWTAuth()` chi mengekstrak claim `sub` via `extractUserIDFromClaims` ke context key `"user_id"` (int), sama seperti `SuccessHandler` echo-jwt lama.
3. **Role validator Kafka + gRPC** — `RoleValidator` (Kafka request-response) dan `RoleValidatorGRPC`/`RequireRoles` di-rewrite ke chi middleware; chain per-route echo `m1(m2(h.X))` → `routerX.With(m1, m2).Get(...)`.
4. **Multipart / file upload** — handler `product`, `category`, `slider`, `merchant_document`, `review_detail` memakai `r.FormValue`/`r.FormFile` (3 nilai) dan `pkg/upload_image.ProcessImageUpload(w, uploadDir, file, isDocument)` (signature 4-argumen khas ecommerce, bukan 2-argumen seperti POS/monolith).
5. **Prefix query/command terpisah** — sebagian besar paket memakai prefix berbeda (`/api/<domain>-query` vs `-command`) sehingga aman dari Route() duplikat; hanya `/api/order-item` yang dobel mount.
6. **Fitur khas microservice dipertahankan** — `pkgresilience` (dependency guard), kafka, redis readiness, pyroscope, format JSON logger echo (field-layout identik).

## 3. Dependensi

**Ditambah:** `go-chi/chi/v5 v5.3.2`, `go-chi/cors v1.2.2`, `swaggo/http-swagger v1.3.4`.
**Tidak lagi di-import langsung:** `labstack/echo/v4`, `echo-jwt/v4`, `echo-swagger` (masih *indirect* via `shared/errors` yang dipakai modul `tests/`).
**Catatan proses:** repo `MamangRust/*` privat sehingga `go get` gagal; require ditambah via `go mod edit` + entri `go.sum` disalin dari monolith, dan build di-resolve lewat `go.work` (`use ./pkg`, `./shared`).

## 4. Struktur Baru

```
service/apigateway/
├── apierror/          # NEW — ApiHandler versi net/http (AppError + HTTPError aware)
├── httpx/             # NEW — JSON/Bind/BindForm/RealIP, request-scoped values, error adapter
├── apps/client.go     # bootstrap chi (Trace → Metrics → Recover → ... → JWT) + http.Server
├── handler/           # 21 subpackage + handler.go + gateway_test.go → chi.Router
├── middlewares/       # auth, role(+gRPC), requiredRole, ratelimit, metrics, trace, pyroscope, responsewriter
└── ...
pkg/upload_image/      # ProcessImageUpload(w, uploadDir, file, isDocument)
tests/                 # 37 file harness terkonversi
```

## 5. Verifikasi

```bash
cd microservice-ecommerce-grpc/service/apigateway
go build ./... && go vet ./...   # ✅
go vet ./handler/...             # ✅ (termasuk kompilasi gateway_test.go)
cd ../../tests && go build ./... && go vet ./... && go test ./...  # lihat hasil di bawah
```

- `handler/gateway_test.go` di-rewrite ke `chi.Walk` untuk inventory rute + normalisasi trailing slash, test swagger↔rute, dan smoke 401/404/400/503 (set test case dipertahankan 1:1).
- Harness `tests/` (37 file): `Deps{E: …}` → `{Router: …}`, bypass middleware `c.Set("user_id", …)` → `context.WithValue`, header konstanta echo → literal string, `MockImageUpload` ikut signature baru, semua `RegisterErrorHandler(...)` dihapus.
- `shared` dan seluruh service gRPC tidak tersentuh; `pkg/upload_image` diubah karena satu-satunya konsumen adalah apigateway.

## 6. Catatan Perilaku

1. **Respons error identik** — `*AppError` tetap dipetakan ke status asli (400/401/403/404/409/429/503/504) dengan trace_id; fallback `echo.NewHTTPError` → `{"message": ...}` direproduksi oleh `httpx.WriteHTTPError`; 404/405 default echo ditiru via `r.NotFound`/`r.MethodNotAllowed`.
2. **Sintaks path param** — `:id` → `{id}`; `GET("")` → `Get("/")`.
3. **Swagger** — `echoSwagger.WrapHandler` → `httpSwagger.WrapHandler` di `/swagger/*`; blank-import docs (`swag.ReadDoc("swagger")`) dipertahankan; anotasi `@Router` tidak diubah.
4. **Echo masih indirect dependency** via `shared/errors`.

## 7. Follow-up (Opsional)

- [ ] Smoke test `hurl/` atau full stack docker.
- [ ] Migrasi `shared/errors` penuh ke net/http → echo hilang total dari repo (juga berlaku untuk `microservice-payment-gateway-grpc` yang belum dimigrasi).
