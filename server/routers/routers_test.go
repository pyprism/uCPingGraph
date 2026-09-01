package routers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestRequestLoggerLogsAtErrorLevelWhenHandlerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	router := gin.New()
	router.Use(requestLogger(zap.New(core)))
	router.GET("/boom", func(c *gin.Context) {
		_ = c.Error(errors.New("handler failed"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "boom"})
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", logs.Len())
	}
	entry := logs.All()[0]
	if entry.Level != zap.ErrorLevel {
		t.Fatalf("expected error-level log for handler error, got %s", entry.Level)
	}
}

func routersTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Network{}, &models.Device{}, &models.Stat{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	return db
}

func TestNewRouterHealthzNilDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.SetDB(nil)
	t.Chdir("..")

	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 with nil DB, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNewRouterHealthzOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.SetDB(routersTestDB(t))
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with live DB, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNewRouterHealthzPingFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := routersTestDB(t)
	models.SetDB(db)
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 with unpingable DB, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNewRouterSetsReleaseModeWithoutDebugEnv(t *testing.T) {
	t.Setenv("DEBUG", "")
	models.SetDB(routersTestDB(t))
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	_ = NewRouter()

	if gin.Mode() != gin.ReleaseMode {
		t.Fatalf("expected release mode when DEBUG is unset, got %s", gin.Mode())
	}
}

func TestNewRouterKeepsDebugModeWhenDebugEnvSet(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	t.Setenv("DEBUG", "True")
	models.SetDB(routersTestDB(t))
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	_ = NewRouter()

	if gin.Mode() != gin.DebugMode {
		t.Fatalf("expected debug mode when DEBUG=True, got %s", gin.Mode())
	}
}

func TestNewRouterServesAPIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.SetDB(routersTestDB(t))
	t.Cleanup(func() { models.SetDB(nil) })
	t.Chdir("..")

	router := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/networks, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"items"`) {
		t.Fatalf("expected items key in response, got %s", rec.Body.String())
	}
}
