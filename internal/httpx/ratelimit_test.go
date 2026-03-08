package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_Allow_BurstThenReject(t *testing.T) {
	rl := NewRateLimiter("test", 1, 3)
	for i := 0; i < 3; i++ {
		if !rl.allow("a") {
			t.Fatalf("request %d should be allowed within burst", i)
		}
	}
	if rl.allow("a") {
		t.Fatal("4th request must be rejected after burst is exhausted")
	}
	if !rl.allow("b") {
		t.Fatal("a fresh key must be allowed independently")
	}
}

func TestRateLimiter_Middleware_429(t *testing.T) {
	rl := NewRateLimiter("test", 1, 1)
	mw := rl.Middleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first call: expected 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second call: expected 429, got %d", second.Code)
	}
}
