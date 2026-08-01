package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInitRejectsInvalidQPS(t *testing.T) {
	if err := Init(Config{GoodsListQPS: 0}); err == nil {
		t.Fatal("Init() error = nil, want error")
	}
}

func TestMiddlewareRejectsRequestsOverLimit(t *testing.T) {
	if err := Init(Config{GoodsListQPS: 1}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/goods", Middleware(GoodsListResource), func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest("GET", "/goods", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", first.Code, http.StatusOK)
	}

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest("GET", "/goods", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
	if second.Header().Get("Retry-After") != "1" {
		t.Fatalf("Retry-After = %q, want %q", second.Header().Get("Retry-After"), "1")
	}
}
