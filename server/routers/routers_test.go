package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRequestLoggerUsesExplicitRequestLatencyField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	router := gin.New()
	router.Use(requestLogger(zap.New(core)))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if logs.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", logs.Len())
	}

	context := logs.All()[0].ContextMap()
	if _, ok := context["request_latency_seconds"]; !ok {
		t.Fatalf("expected request_latency_seconds field, got %#v", context)
	}
	if _, ok := context["latency"]; ok {
		t.Fatalf("did not expect ambiguous latency field, got %#v", context)
	}
}
