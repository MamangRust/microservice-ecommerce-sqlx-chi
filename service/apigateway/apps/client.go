package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/handler"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/middlewares"
	// Blank import: registers the generated Swagger spec (docs.init → swag.Register)
	// so http-swagger's /swagger/doc.json can resolve swag.ReadDoc("swagger").
	_ "github.com/MamangRust/microservice-ecommerce-grpc-apigateway/docs"
	"github.com/MamangRust/microservice-ecommerce-pkg/auth"
	"github.com/MamangRust/microservice-ecommerce-pkg/dotenv"
	"github.com/MamangRust/microservice-ecommerce-pkg/kafka"
	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	pkgmiddleware "github.com/MamangRust/microservice-ecommerce-pkg/middleware"
	otel_pkg "github.com/MamangRust/microservice-ecommerce-pkg/otel"
	"github.com/MamangRust/microservice-ecommerce-pkg/resilience"
	"github.com/MamangRust/microservice-ecommerce-pkg/upload_image"
	"github.com/MamangRust/microservice-ecommerce-shared/cache"
	"github.com/MamangRust/microservice-ecommerce-shared/observability"
	"github.com/go-chi/chi/v5"
	chi_mw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/grafana/pyroscope-go"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// readinessRedis holds the gateway Redis client for the /ready endpoint.
// It is set once Redis is initialised during client bootstrap.
var readinessRedis *redis.Client

const (
	defaultServerPort             = ":5000"
	defaultWindowSizeClient       = 16 * 1024 * 1024
	defaultKeepaliveTimeClient    = 20 * time.Second
	defaultKeepaliveTimeoutClient = 5 * time.Second

	monitoringInterval     = 30 * time.Second
	cleanupInterval        = 5 * time.Minute
	cacheRefCountThreshold = 10000
	shutdownTimeout        = 30 * time.Second

	redisDialTimeout  = 5 * time.Second
	redisReadTimeout  = 3 * time.Second
	redisWriteTimeout = 3 * time.Second
	redisPoolSize     = 10
	redisMinIdleConns = 5

	// authRateLimitRPS/Burst guard the credential endpoints against brute force.
	authRateLimitRPS   = 10
	authRateLimitBurst = 10
)

// @title Ecommerce gRPC
// @version 1.0
// @description gRPC based Ecommerce service
// @host localhost:5000
// @BasePath /api/
// @securityDefinitions.apikey BearerAuth
// @in Header
// @name Authorization

// Client represents the main application client
type Client struct {
	Logger       logger.LoggerInterface
	Router       chi.Router
	Server       *http.Server
	GRPCConn     *grpc.ClientConn
	TokenManager *auth.Manager
	Telemetry    *otel_pkg.Telemetry
	Config       *ClientConfig
	Redis        *redis.Client
	cancelTasks  context.CancelFunc
	tasksDone    []<-chan struct{}
}

type ClientConfig struct {
	ServiceName          string  `mapstructure:"service_name"`
	ServiceVersion       string  `mapstructure:"service_version"`
	Environment          string  `mapstructure:"environment"`
	OtelEndpoint         string  `mapstructure:"otel_endpoint"`
	OtelSamplingFraction float64 `mapstructure:"otel_sampling_fraction"`
	ServerPort           string  `mapstructure:"server_port"`

	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type CacheManager struct {
	cache  *cache.CacheStore
	logger logger.LoggerInterface
}

// ServiceAddresses maps gRPC service endpoints.
//
// Note on configuration: loadServiceAddresses reads these from the
// GRPC_<NAME>_ADDR environment variables (e.g. GRPC_PRODUCT_ADDR) with
// fallbacks to the defaults below. It deliberately does NOT rely on
// viper.SetEnvPrefix: dotenv.Viper() already called AutomaticEnv() without a
// prefix during bootstrap, so a later SetEnvPrefix has no effect on env
// lookups. Reading each variable explicitly avoids that pitfall.
type ServiceAddresses struct {
	Auth             string `mapstructure:"auth"`
	Role             string `mapstructure:"role"`
	User             string `mapstructure:"user"`
	Category         string `mapstructure:"category"`
	Merchant         string `mapstructure:"merchant"`
	OrderItem        string `mapstructure:"order_item"`
	Order            string `mapstructure:"order"`
	Product          string `mapstructure:"product"`
	Transaction      string `mapstructure:"transaction"`
	Cart             string `mapstructure:"cart"`
	Review           string `mapstructure:"review"`
	Slider           string `mapstructure:"slider"`
	Shipping         string `mapstructure:"shipping"`
	Banner           string `mapstructure:"banner"`
	MerchantAward    string `mapstructure:"merchant_award"`
	MerchantBusiness string `mapstructure:"merchant_business"`
	MerchantDetail   string `mapstructure:"merchant_detail"`
	MerchantPolicy   string `mapstructure:"merchant_policy"`
	ReviewDetail     string `mapstructure:"review_detail"`
}

func NewCacheManager(cache *cache.CacheStore, logger logger.LoggerInterface) *CacheManager {
	return &CacheManager{
		cache:  cache,
		logger: logger,
	}
}

func (cm *CacheManager) StartMonitoring(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(monitoringInterval)
		defer ticker.Stop()

		cm.logger.Info("Cache monitoring task started",
			zap.Duration("interval", monitoringInterval),
		)

		for {
			select {
			case <-ctx.Done():
				cm.logger.Info("Cache monitoring task stopped")
				return
			case <-ticker.C:
				cm.monitor(ctx)
			}
		}
	}()

	return done
}

func (cm *CacheManager) monitor(ctx context.Context) {
	refCount := cm.cache.GetRefCount()

	stats, err := cm.cache.GetStats(ctx)
	if err != nil {
		cm.logger.Error("Failed to get cache stats", zap.Error(err))
		return
	}

	logLevel := zap.InfoLevel
	if refCount > cacheRefCountThreshold {
		logLevel = zap.WarnLevel
	}

	if ce := cm.logger.Check(logLevel, "Cache statistics"); ce != nil {
		ce.Write(
			zap.Int64("ref_count", refCount),
			zap.Int64("total_keys", stats.TotalKeys),
			zap.Float64("hit_rate", stats.HitRate),
			zap.String("memory_used", stats.MemoryUsedHuman),
			zap.Bool("high_ref_count", refCount > cacheRefCountThreshold),
		)
	}
}

func (cm *CacheManager) StartCleanup(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		cm.logger.Info("Cache cleanup task started",
			zap.Duration("interval", cleanupInterval),
		)

		for {
			select {
			case <-ctx.Done():
				cm.logger.Info("Cache cleanup task stopped")
				return
			case <-ticker.C:
				cm.cleanup(ctx)
			}
		}
	}()

	return done
}

func (cm *CacheManager) cleanup(ctx context.Context) {
	cm.logger.Info("Starting periodic cache cleanup")

	statsBefore, err := cm.cache.GetStats(ctx)
	if err != nil {
		cm.logger.Error("Failed to get cache stats before cleanup", zap.Error(err))
		statsBefore = nil
	}

	scanned, err := cm.cache.ClearExpired(ctx)
	if err != nil {
		cm.logger.Error("Cache cleanup failed", zap.Error(err))
		return
	}

	statsAfter, err := cm.cache.GetStats(ctx)
	if err != nil {
		cm.logger.Error("Failed to get cache stats after cleanup", zap.Error(err))
		statsAfter = nil
	}

	logFields := []zap.Field{
		zap.Int64("scanned_keys", scanned),
		zap.Int64("ref_count", cm.cache.GetRefCount()),
	}

	if statsBefore != nil && statsAfter != nil {
		keysRemoved := statsBefore.TotalKeys - statsAfter.TotalKeys
		logFields = append(logFields,
			zap.Int64("keys_before", statsBefore.TotalKeys),
			zap.Int64("keys_after", statsAfter.TotalKeys),
			zap.Int64("keys_removed", keysRemoved),
			zap.String("memory_before", statsBefore.MemoryUsedHuman),
			zap.String("memory_after", statsAfter.MemoryUsedHuman),
		)
	}

	cm.logger.Info("Cache cleanup completed", logFields...)
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	if err := dotenv.Viper(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	if err := initPyroscope(); err != nil {
		log.Fatal("Failed to initialize pyroscope:", err)
	}

	cfg, err := loadClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	telemetry, err := initTelemetry(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	cacheMetrics, err := observability.NewCacheMetrics("cache")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache metrics: %w", err)
	}

	logger, err := logger.NewLogger(cfg.ServiceName, telemetry.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	tokenManager, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		return nil, fmt.Errorf("failed to create token manager: %w", err)
	}

	addresses, err := loadServiceAddresses()

	if err != nil {
		return nil, fmt.Errorf("failed to load service")
	}

	// One shared interceptor across every downstream connection: each gRPC
	// method gets its own breaker/bulkhead (keyed by full method name), so a
	// slow or failing service cannot exhaust the gateway's goroutines or open
	// the circuit for unrelated services.
	guard := resilience.NewDependencyGuardInterceptor(logger)

	conns, err := createServiceConnections(addresses, logger, guard)
	if err != nil {
		return nil, fmt.Errorf("failed to connect services: %w", err)
	}

	router, server := createChiServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisClient, err := initRedisClient(ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	readinessRedis = redisClient

	myKafka := kafka.NewKafka(logger, []string{viper.GetString("KAFKA_BROKERS")})

	cacheStore := cache.NewCacheStore(redisClient, logger, cacheMetrics)

	tasksCtx, cancelTasks := context.WithCancel(context.Background())
	cacheManager := NewCacheManager(cacheStore, logger)

	tasksDone := []<-chan struct{}{
		cacheManager.StartMonitoring(tasksCtx),
		cacheManager.StartCleanup(tasksCtx),
	}

	handlerDeps := &handler.Deps{
		Kafka:              myKafka,
		ServiceConnections: conns,
		Token:              tokenManager,
		Router:             router,
		Logger:             logger,
		Cache:              cacheStore,
		Image:              upload_image.NewImageUpload(logger),
	}
	handler.NewHandler(handlerDeps)

	client := &Client{
		Logger:       logger,
		Router:       router,
		Server:       server,
		TokenManager: tokenManager,
		Telemetry:    telemetry,
		Config:       cfg,
		Redis:        redisClient,
		cancelTasks:  cancelTasks,
		tasksDone:    tasksDone,
	}

	logger.Info("Client initialized successfully",
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.ServiceVersion),
		zap.String("server_port", cfg.ServerPort),
	)

	return client, nil
}

func (c *Client) gracefulShutdown() error {
	c.Logger.Info("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := c.Server.Shutdown(ctx); err != nil {
		c.Logger.Error("HTTP server shutdown error", zap.Error(err))
		return fmt.Errorf("failed to shutdown http server: %w", err)
	}

	c.Logger.Info("HTTP server stopped gracefully")
	return nil
}

func (c *Client) Cleanup() {
	c.Logger.Info("Cleaning up resources...")

	if c.cancelTasks != nil {
		c.Logger.Info("Stopping background tasks")
		c.cancelTasks()

		for i, done := range c.tasksDone {
			c.Logger.Debug("Waiting for background task to complete", zap.Int("task_index", i))
			<-done
		}
		c.Logger.Info("All background tasks stopped")
	}

	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.Logger.Error("Failed to close Redis connection", zap.Error(err))
		} else {
			c.Logger.Info("Redis connection closed")
		}
	}

	if c.GRPCConn != nil {
		if err := c.GRPCConn.Close(); err != nil {
			c.Logger.Error("Failed to close gRPC connection", zap.Error(err))
		} else {
			c.Logger.Info("gRPC connection closed")
		}
	}

	if c.Telemetry != nil {
		if err := c.Telemetry.Shutdown(context.Background()); err != nil {
			c.Logger.Error("Failed to shutdown telemetry", zap.Error(err))
		} else {
			c.Logger.Info("Telemetry shutdown successfully")
		}
	}

	if c.Logger != nil {
		_ = c.Logger.Sync()
	}

	c.Logger.Info("Cleanup completed")
}

func initPyroscope() error {
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "apigateway",
		ServerAddress:   os.Getenv("PYROSCOPE_SERVER"),
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
		Tags: map[string]string{
			"service": "apigateway",
			"env":     os.Getenv("ENV"),
			"version": os.Getenv("VERSION"),
		},
	})
	return err
}

func initTelemetry(cfg *ClientConfig) (*otel_pkg.Telemetry, error) {
	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:            cfg.ServiceName,
		ServiceVersion:         cfg.ServiceVersion,
		Environment:            cfg.Environment,
		Endpoint:               cfg.OtelEndpoint,
		Insecure:               true,
		EnableRuntimeMetrics:   true,
		RuntimeMetricsInterval: 15 * time.Second,
		SamplingFraction:       cfg.OtelSamplingFraction,
	})

	if err := telemetry.Init(context.Background()); err != nil {
		return nil, err
	}

	return telemetry, nil
}

func initRedisClient(ctx context.Context, logger logger.LoggerInterface) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST_APIGATEWAY"), viper.GetString("REDIS_PORT_APIGATEWAY")),
		Password:     viper.GetString("REDIS_PASSWORD_APIGATEWAY"),
		DB:           viper.GetInt("REDIS_DB_APIGATEWAY"),
		DialTimeout:  redisDialTimeout,
		ReadTimeout:  redisReadTimeout,
		WriteTimeout: redisWriteTimeout,
		PoolSize:     redisPoolSize,
		MinIdleConns: redisMinIdleConns,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis connection established",
		zap.String("addr", fmt.Sprintf("%s:%s", viper.GetString("REDIS_HOST_APIGATEWAY"), viper.GetString("REDIS_PORT_APIGATEWAY"))),
		zap.Int("db", viper.GetInt("REDIS_DB_APIGATEWAY")),
	)

	return client, nil
}

func createChiServer(cfg *ClientConfig) (chi.Router, *http.Server) {
	r := chi.NewRouter()

	httpMetrics, err := middlewares.NewHTTPMetrics(cfg.ServiceName)
	if err != nil {
		log.Println("Failed to initialize HTTP metrics middleware")
	}

	// Trace must be the outermost middleware so metrics and handlers record
	// within the traced request context (end-to-end trace linkage).
	r.Use(middlewares.TraceMiddleware(cfg.ServiceName))
	if httpMetrics != nil {
		r.Use(httpMetrics.Middleware())
	}
	r.Use(chi_mw.Recoverer)
	r.Use(chi_mw.RequestID)
	r.Use(createLoggerMiddleware())
	r.Use(middlewares.PyroscopeMiddleware())
	r.Use(createCORSMiddleware(cfg.AllowedOrigins))
	r.Use(chi_mw.Compress(5))
	r.Use(createSecureMiddleware())

	// Rate limit credential endpoints (login/register) before authentication so
	// brute-force attempts receive 429 instead of reaching the auth service.
	authRateLimiter := middlewares.NewRateLimiter(authRateLimitRPS, authRateLimitBurst)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/auth/") {
				authRateLimiter.Limit(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Use(middlewares.JWTAuth())

	// Mirror echo's default responses for unmatched routes and methods so API
	// clients keep receiving the same compact JSON error bodies.
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.JSON(w, http.StatusNotFound, map[string]string{"message": "Not Found"})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.JSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "Method Not Allowed"})
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", createHealthHandler(cfg))
	r.Get("/ready", createReadinessHandler(cfg))

	server := &http.Server{
		Addr:              cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return r, server
}

func createHealthHandler(cfg *ClientConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "healthy",
			"service": cfg.ServiceName,
			"version": cfg.ServiceVersion,
			"time":    time.Now().UTC(),
		})
	}
}

// createReadinessHandler reports whether the gateway is ready to serve traffic.
// Unlike the liveness /health endpoint, it performs lightweight dependency
// checks (Redis ping) so load balancers/orchestrators can distinguish a running
// process from one that is actually able to serve requests.
func createReadinessHandler(cfg *ClientConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if readinessRedis == nil {
			_ = httpx.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"deps": map[string]string{
					"redis": "not_configured",
				},
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readinessRedis.Ping(ctx).Err(); err != nil {
			_ = httpx.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"deps": map[string]string{
					"redis": "down",
				},
			})
			return
		}

		_ = httpx.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  "ready",
			"service": cfg.ServiceName,
			"version": cfg.ServiceVersion,
			"deps": map[string]string{
				"redis": "up",
			},
			"time": time.Now().UTC(),
		})
	}
}

// jsonLogFormatter renders one JSON line per request, keeping the same field
// layout as the former echo LoggerWithConfig output.
type jsonLogFormatter struct{}

type jsonLogEntry struct {
	start    time.Time
	reqID    string
	remoteIP string
	host     string
	method   string
	uri      string
	bytesIn  int64
}

func (f jsonLogFormatter) NewLogEntry(r *http.Request) chi_mw.LogEntry {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return &jsonLogEntry{
		start:    time.Now(),
		reqID:    chi_mw.GetReqID(r.Context()),
		remoteIP: httpx.RealIP(r),
		host:     r.Host,
		method:   r.Method,
		uri:      fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI),
		bytesIn:  r.ContentLength,
	}
}

func (e *jsonLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	payload := map[string]interface{}{
		"time":          time.Now().Format(time.RFC3339),
		"id":            e.reqID,
		"remote_ip":     e.remoteIP,
		"host":          e.host,
		"method":        e.method,
		"uri":           e.uri,
		"status":        status,
		"error":         "",
		"latency":       elapsed.Nanoseconds(),
		"latency_human": elapsed.String(),
		"bytes_in":      e.bytesIn,
		"bytes_out":     bytes,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

func (e *jsonLogEntry) Panic(value interface{}, stack []byte) {
	// Panics are handled and logged by chi_mw.Recoverer.
}

func createLoggerMiddleware() func(http.Handler) http.Handler {
	return chi_mw.RequestLogger(jsonLogFormatter{})
}

func createCORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

func createSecureMiddleware() func(http.Handler) http.Handler {
	csp := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; " +
		"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; " +
		"img-src 'self' data: https:; " +
		"font-src 'self' data: https://cdnjs.cloudflare.com; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none';"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Set("X-XSS-Protection", "1; mode=block")
			header.Set("X-Content-Type-Options", "nosniff")
			header.Set("X-Frame-Options", "SAMEORIGIN")
			header.Set("Strict-Transport-Security", "max-age=31536000; includeSubdomains; preload")
			header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			header.Set("Content-Security-Policy", csp)
			next.ServeHTTP(w, r)
		})
	}
}

func loadClientConfig() (*ClientConfig, error) {
	v := viper.GetViper()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("service_name", "apigateway")
	v.SetDefault("service_version", "1.0.0")
	v.SetDefault("environment", "production")
	v.SetDefault("otel_endpoint", "otel-collector:4317")
	v.SetDefault("otel_sampling_fraction", 1.0)
	v.SetDefault("server_port", defaultServerPort)

	v.SetDefault("allowed_origins", []string{"http://localhost:1420"})

	var cfg ClientConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client config: %w", err)
	}

	return &cfg, nil
}

func loadServiceAddresses() (*ServiceAddresses, error) {
	v := viper.GetViper()

	// Read GRPC_<NAME>_ADDR explicitly. dotenv.Viper() already enabled
	// AutomaticEnv() without a prefix, so SetEnvPrefix here would be a no-op;
	// reading each env var directly is the reliable approach.
	getAddr := func(name, fallback string) string {
		if addr := v.GetString("GRPC_" + name + "_ADDR"); addr != "" {
			return addr
		}
		return fallback
	}

	cfg := &ServiceAddresses{
		Auth:             getAddr("AUTH", "auth:50051"),
		Role:             getAddr("ROLE", "role:50052"),
		User:             getAddr("USER", "user:50053"),
		Category:         getAddr("CATEGORY", "category:50054"),
		Merchant:         getAddr("MERCHANT", "merchant:50055"),
		OrderItem:        getAddr("ORDER_ITEM", "order-item:50056"),
		Order:            getAddr("ORDER", "order:50057"),
		Product:          getAddr("PRODUCT", "product:50058"),
		Transaction:      getAddr("TRANSACTION", "transaction:50059"),
		Cart:             getAddr("CART", "cart:50060"),
		Review:           getAddr("REVIEW", "review:50061"),
		Slider:           getAddr("SLIDER", "slider:50062"),
		Shipping:         getAddr("SHIPPING", "shipping_address:50063"),
		Banner:           getAddr("BANNER", "banner:50064"),
		MerchantAward:    getAddr("MERCHANT_AWARD", "merchant_award:50065"),
		MerchantBusiness: getAddr("MERCHANT_BUSINESS", "merchant_business:50066"),
		MerchantDetail:   getAddr("MERCHANT_DETAIL", "merchant_detail:50067"),
		MerchantPolicy:   getAddr("MERCHANT_POLICY", "merchant_policy:50068"),
		ReviewDetail:     getAddr("REVIEW_DETAIL", "review_detail:50069"),
	}

	return cfg, nil
}

func createServiceConnections(addresses *ServiceAddresses, logger logger.LoggerInterface, guard *resilience.DependencyGuardInterceptor) (*handler.ServiceConnections, error) {
	connections := &handler.ServiceConnections{}

	serviceMap := map[string]struct {
		addr *string
		conn **grpc.ClientConn
	}{
		"Auth":             {&addresses.Auth, &connections.Auth},
		"Role":             {&addresses.Role, &connections.Role},
		"User":             {&addresses.User, &connections.User},
		"Category":         {&addresses.Category, &connections.Category},
		"Merchant":         {&addresses.Merchant, &connections.Merchant},
		"OrderItem":        {&addresses.OrderItem, &connections.OrderItem},
		"Order":            {&addresses.Order, &connections.Order},
		"Product":          {&addresses.Product, &connections.Product},
		"Transaction":      {&addresses.Transaction, &connections.Transaction},
		"Cart":             {&addresses.Cart, &connections.Cart},
		"Review":           {&addresses.Review, &connections.Review},
		"Slider":           {&addresses.Slider, &connections.Slider},
		"Shipping":         {&addresses.Shipping, &connections.Shipping},
		"Banner":           {&addresses.Banner, &connections.Banner},
		"MerchantAward":    {&addresses.MerchantAward, &connections.MerchantAward},
		"MerchantBusiness": {&addresses.MerchantBusiness, &connections.MerchantBusiness},
		"MerchantDetail":   {&addresses.MerchantDetail, &connections.MerchantDetail},
		"MerchantPolicy":   {&addresses.MerchantPolicy, &connections.MerchantPolicy},
		"ReviewDetail":     {&addresses.ReviewDetail, &connections.ReviewDetail},
	}

	for name, svc := range serviceMap {
		conn, err := createConnection(*svc.addr, name, logger, guard)
		if err != nil {
			closeConnections(connections, logger)
			return nil, err
		}
		*svc.conn = conn
	}

	// merchant service meng-host gRPC merchant_document pada listener yang sama.
	// Tanpa alias ini ServiceConnections.MerchantDocument nil dan handler panic.
	if connections.Merchant != nil {
		connections.MerchantDocument = connections.Merchant
	}

	return connections, nil
}

func createConnection(address, serviceName string, logger logger.LoggerInterface, guard *resilience.DependencyGuardInterceptor) (*grpc.ClientConn, error) {
	logger.Info(fmt.Sprintf("Connecting to %s service", serviceName),
		zap.String("address", address),
	)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithInitialConnWindowSize(defaultWindowSizeClient),
		grpc.WithInitialWindowSize(defaultWindowSizeClient),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                defaultKeepaliveTimeClient,
			Timeout:             defaultKeepaliveTimeoutClient,
			PermitWithoutStream: true,
		}),
		// F6: gateway resilience (per-call deadline + circuit breaker + bulkhead)
		// keyed per gRPC method, then propagate the current trace context to
		// downstream gRPC services so a single request can be traced end-to-end
		// from the gateway to the domain.
		grpc.WithChainUnaryInterceptor(
			guard.UnaryInterceptor(),
			pkgmiddleware.TraceUnaryClientInterceptor(),
		),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service", serviceName), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to %s service: %w", serviceName, err)
	}

	logger.Info(fmt.Sprintf("Successfully connected to %s service", serviceName))
	return conn, nil
}

func closeConnections(conns *handler.ServiceConnections, logger logger.LoggerInterface) {
	connectionMap := map[string]*grpc.ClientConn{
		"Auth":        conns.Auth,
		"Role":        conns.Role,
		"Card":        conns.Card,
		"Merchant":    conns.Merchant,
		"User":        conns.User,
		"Saldo":       conns.Saldo,
		"Topup":       conns.Topup,
		"Transaction": conns.Transaction,
		"Transfer":    conns.Transfer,
		"Withdraw":    conns.Withdraw,
	}

	for name, conn := range connectionMap {
		if conn != nil {
			if err := conn.Close(); err != nil {
				logger.Error(fmt.Sprintf("Failed to close %s connection", name), zap.Error(err))
			} else {
				logger.Info(fmt.Sprintf("%s connection closed", name))
			}
		}
	}
}

func RunClient() (*Client, func(), error) {
	client, err := NewClient(nil)
	if err != nil {
		return nil, nil, err
	}

	// Start the HTTP server. Previously RunClient never started the server, so
	// the process sat idle in main() waiting for a signal while nothing bound
	// ServerPort. Keep signal handling and cleanup in main() (via shutdown) to
	// avoid double-Cleanup races with Run()'s own signal handlers.
	go func() {
		client.Logger.Info("HTTP server starting",
			zap.String("port", client.Config.ServerPort),
			zap.String("swagger", "http://localhost"+client.Config.ServerPort+"/swagger/index.html"),
		)
		if err := client.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			client.Logger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()

	shutdown := func() {
		if err := client.gracefulShutdown(); err != nil {
			client.Logger.Error("Graceful shutdown failed", zap.Error(err))
		}
		client.Cleanup()
	}

	return client, shutdown, nil
}
