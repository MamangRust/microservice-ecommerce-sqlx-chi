package middlewares

import (
	"net/http"

	"github.com/MamangRust/microservice-ecommerce-grpc-apigateway/httpx"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return &RateLimiter{
		limiter: limiter,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.limiter.Allow() {
			_ = httpx.JSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many requests, please try again later",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
