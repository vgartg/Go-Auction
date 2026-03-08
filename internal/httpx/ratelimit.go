package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/vgartg/goauction/internal/metrics"
)

type RateLimiter struct {
	rate     float64
	burst    float64
	gcAfter  time.Duration
	scope    string
	mu       sync.Mutex
	buckets  map[string]*bucket
	lastSeen map[string]time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewRateLimiter(scope string, ratePerSecond, burst float64) *RateLimiter {
	rl := &RateLimiter{
		rate:     ratePerSecond,
		burst:    burst,
		gcAfter:  5 * time.Minute,
		scope:    scope,
		buckets:  make(map[string]*bucket),
		lastSeen: make(map[string]time.Time),
	}
	go rl.gcLoop()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	rl.lastSeen[key] = now
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.burst - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(rl.burst, b.tokens+elapsed*rl.rate)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *RateLimiter) gcLoop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for range t.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.gcAfter)
		for k, ls := range rl.lastSeen {
			if ls.Before(cutoff) {
				delete(rl.buckets, k)
				delete(rl.lastSeen, k)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientIP(r)
			if !rl.allow(key) {
				metrics.RateLimitedTotal.WithLabelValues(rl.scope).Inc()
				w.Header().Set("Retry-After", "1")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
