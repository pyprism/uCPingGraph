package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pyprism/uCPingGraph/logger"
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupControllerTestDB(t *testing.T) string {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Network{}, &models.Device{}, &models.Stat{}); err != nil {
		t.Fatalf("auto-migrate test db: %v", err)
	}
	models.SetDB(db)

	network := models.Network{}
	networkID, err := network.CreateNetwork("home")
	if err != nil {
		t.Fatalf("create network: %v", err)
	}

	device := models.Device{}
	_, token, err := device.CreateDevice(int(networkID), "esp32-main")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}
	return token
}

func TestPostStatsRequiresAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(`{"latency_ms":10,"sent_packets":1,"received_packets":1,"packet_loss_percent":0}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPostStatsCreatesRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	body := `{"latency_ms":22.4,"sent_packets":5,"received_packets":4,"packet_loss_percent":20,"target":"1.1.1.1","platform":"esp8266","rssi":-70}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var count int64
	if err := models.DB.Model(&models.Stat{}).Count(&count).Error; err != nil {
		t.Fatalf("count stats: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stat record, got %d", count)
	}
}

func TestPostStatsLogsESPReportedLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	core, logs := observer.New(zap.InfoLevel)
	logger.L = zap.New(core)
	t.Cleanup(func() {
		logger.L = zap.NewNop()
	})
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	body := `{"latency_ms":22.4,"sent_packets":5,"received_packets":4,"packet_loss_percent":20,"target":"1.1.1.1","platform":"esp8266","rssi":-70}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	if logs.Len() != 1 {
		t.Fatalf("expected 1 log entry, got %d", logs.Len())
	}
	entry := logs.All()[0]
	if entry.Message != "esp telemetry received" {
		t.Fatalf("unexpected log message %q", entry.Message)
	}
	context := entry.ContextMap()
	if context["esp_ping_latency_ms"] != 22.4 {
		t.Fatalf("expected esp_ping_latency_ms=22.4, got %#v", context["esp_ping_latency_ms"])
	}
	if context["packet_loss_percent"] != 20.0 {
		t.Fatalf("expected packet_loss_percent=20, got %#v", context["packet_loss_percent"])
	}
}

func TestPostStatsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostStatsNegativeLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	body := `{"latency_ms":-5,"sent_packets":1,"received_packets":1,"packet_loss_percent":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative latency, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostStatsMissingLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	body := `{"sent_packets":1,"received_packets":0,"packet_loss_percent":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing latency, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostStatsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	body := `{"latency_ms":10,"sent_packets":1,"received_packets":1,"packet_loss_percent":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "invalid-token-here")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", rec.Code)
	}
}

func TestPostStatsLegacyLatencyField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	// Old clients send "latency" instead of "latency_ms"
	body := `{"latency":15.5}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for legacy format, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostStatsBackwardCompatZeroPackets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	// Old client that only sends latency_ms, no packet info
	body := `{"latency_ms":12.3}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for backward compat, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostStatsPacketLossWithoutCounters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.POST("/api/stats", api.PostStats)

	// packet_loss_percent supplied but sent/received counters absent.
	body := `{"latency_ms":12,"packet_loss_percent":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/stats", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestNetworks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/networks", api.Networks)

	req := httptest.NewRequest(http.MethodGet, "/api/networks", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string][]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	items := resp["items"]
	if len(items) != 1 || items[0] != "home" {
		t.Fatalf("expected [home], got %v", items)
	}
}

func TestDevicesByNetwork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/networks/:network/devices", api.DevicesByNetwork)

	req := httptest.NewRequest(http.MethodGet, "/api/networks/home/devices", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string][]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp["items"]) != 1 || resp["items"][0] != "esp32-main" {
		t.Fatalf("expected [esp32-main], got %v", resp["items"])
	}
}

func TestDevicesByNetworkNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/networks/:network/devices", api.DevicesByNetwork)

	req := httptest.NewRequest(http.MethodGet, "/api/networks/nonexistent/devices", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSeriesReturnsSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token := setupControllerTestDB(t)

	if err := service.SaveStats(token, service.IngestTelemetry{
		LatencyMs:         19,
		SentPackets:       5,
		ReceivedPackets:   5,
		PacketLossPercent: 0,
		Target:            "1.1.1.1",
		Platform:          "esp32",
		RSSI:              -52,
	}); err != nil {
		t.Fatalf("seed telemetry: %v", err)
	}

	api := new(APIController)
	router := gin.New()
	router.GET("/api/series", api.Series)

	req := httptest.NewRequest(http.MethodGet, "/api/series?network=home&device=esp32-main&minutes=60", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload models.MetricsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Summary.Samples != 1 {
		t.Fatalf("expected 1 sample, got %d", payload.Summary.Samples)
	}
	if len(payload.Series.LatencySeries) != 1 {
		t.Fatalf("expected one latency datapoint, got %d", len(payload.Series.LatencySeries))
	}
}

func TestSeriesMissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/series", api.Series)

	// Missing both network and device
	req := httptest.NewRequest(http.MethodGet, "/api/series?minutes=60", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSeriesInvalidMinutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/series", api.Series)

	tests := []struct {
		name    string
		minutes string
	}{
		{"negative", "-1"},
		{"zero", "0"},
		{"too_large", "99999"},
		{"non_numeric", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet,
				"/api/series?network=home&device=esp32-main&minutes="+tt.minutes, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for minutes=%s, got %d", tt.minutes, rec.Code)
			}
		})
	}
}

func TestSeriesNotFoundDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupControllerTestDB(t)
	api := new(APIController)

	router := gin.New()
	router.GET("/api/series", api.Series)

	req := httptest.NewRequest(http.MethodGet, "/api/series?network=home&device=ghost&minutes=60", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
