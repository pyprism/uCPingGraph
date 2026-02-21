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
	"github.com/pyprism/uCPingGraph/models"
	"github.com/pyprism/uCPingGraph/service"
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
