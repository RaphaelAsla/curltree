package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"

	"curltree/pkg/utils"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	clients map[string]*rate.Limiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
	logger  *utils.Logger
}

func NewRateLimiter(requestsPerMinute int, burst int, logger *utils.Logger) *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*rate.Limiter),
		rate:    rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:   burst,
		logger:  logger.WithContext("rate_limiter"),
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.clients[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter, exists = rl.clients[ip]
		if !exists {
			limiter = rate.NewLimiter(rl.rate, rl.burst)
			rl.clients[ip] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			rl.logger.LogRateLimit(ip, int(rl.rate))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (rl *RateLimiter) CleanupOldClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, limiter := range rl.clients {
		if limiter.TokensAt(time.Now()) == float64(rl.burst) {
			delete(rl.clients, ip)
		}
	}
}

func (rl *RateLimiter) StartCleanupTask() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			rl.CleanupOldClients()
		}
	}()
}

func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		return xForwardedFor
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}