package resilience

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-ecommerce-pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DependencyGuard wraps calls to a downstream gRPC dependency with three
// protections:
//
//  1. Per-call deadline (context.WithTimeout) so a slow dependency cannot hold
//     a request open indefinitely.
//  2. Bulkhead (RequestLimiter) so a failing dependency cannot exhaust the
//     caller's goroutines/connections.
//  3. Circuit breaker so repeated transport failures stop hammering the
//     dependency and fail fast with Unavailable.
//
// Only transport-level failures (Unavailable / DeadlineExceeded / Aborted /
// ResourceExhausted) count against the circuit breaker. Business errors (not
// found, validation, insufficient balance) are normal responses and must NOT
// open the circuit.
type DependencyGuard struct {
	name        string
	breaker     *CircuitBreaker
	limiter     *RequestLimiter
	callTimeout time.Duration
}

// NewDependencyGuard creates a guard for one downstream dependency. A nil
// logger is replaced by a no-op logger so the guard is safe to construct in
// tests and in code paths without logging.
func NewDependencyGuard(name string, threshold uint64, timeoutSecs uint64, maxConcurrent int64, callTimeout time.Duration, log logger.LoggerInterface) *DependencyGuard {
	if callTimeout <= 0 {
		callTimeout = 3 * time.Second
	}
	if log == nil {
		log = noopLogger{}
	}
	return &DependencyGuard{
		name:        name,
		breaker:     NewCircuitBreaker(threshold, timeoutSecs, log),
		limiter:     NewRequestLimiter(maxConcurrent, log),
		callTimeout: callTimeout,
	}
}

// Call executes fn under the guard. fn receives a context that already carries
// the per-call timeout. The returned error is passed through unchanged.
func (g *DependencyGuard) Call(ctx context.Context, fn func(ctx context.Context) error) error {
	if g == nil {
		return fn(ctx)
	}

	if !g.limiter.TryAcquire() {
		return status.Errorf(codes.ResourceExhausted, "%s dependency bulkhead full (max %d concurrent)", g.name, g.limiter.MaxConcurrent())
	}
	defer g.limiter.Release()

	if !g.breaker.ShouldAllowRequest() {
		return status.Errorf(codes.Unavailable, "%s dependency circuit open; failing fast", g.name)
	}

	callCtx, cancel := context.WithTimeout(ctx, g.callTimeout)
	defer cancel()

	err := fn(callCtx)
	if isTransportFailure(err) {
		g.breaker.RecordFailure()
	} else {
		g.breaker.RecordSuccess()
	}
	return err
}

// isTransportFailure reports whether err is a transport-level dependency
// failure that should count against the circuit breaker.
func isTransportFailure(err error) bool {
	if err == nil {
		return false
	}
	code := status.Code(err)
	return code == codes.Unavailable ||
		code == codes.DeadlineExceeded ||
		code == codes.Aborted ||
		code == codes.ResourceExhausted
}

// noopLogger satisfies logger.LoggerInterface and swallows everything. It is
// used when a guard is constructed without a logger (tests, optional paths).
type noopLogger struct{}

func (noopLogger) Info(string, ...zap.Field)                {}
func (noopLogger) Fatal(string, ...zap.Field)               {}
func (noopLogger) Debug(string, ...zap.Field)               {}
func (noopLogger) Error(string, ...zap.Field)               {}
func (noopLogger) Warn(string, ...zap.Field)                {}
func (noopLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry { return nil }
func (noopLogger) With(...zap.Field) logger.LoggerInterface { return noopLogger{} }
func (noopLogger) Sync() error                              { return nil }
