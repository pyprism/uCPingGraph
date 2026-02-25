package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pyprism/uCPingGraph/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
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
}

func seedDevice(t *testing.T) string {
	t.Helper()
	network := models.Network{}
	networkID, err := network.CreateNetwork("home")
	if err != nil {
		t.Fatalf("create network: %v", err)
	}

	device := models.Device{}
	_, token, err := device.CreateDevice(int(networkID), "esp8266-node")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}

	return token
}

func TestSaveStatsPersistsTelemetry(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(token, IngestTelemetry{
		LatencyMs:         12.5,
		SentPackets:       5,
		ReceivedPackets:   4,
		PacketLossPercent: 20,
		Target:            "1.1.1.1",
		Platform:          "esp32",
		RSSI:              -62,
	})
	if err != nil {
		t.Fatalf("save stats failed: %v", err)
	}

	var stats []models.Stat
	if err := models.DB.Find(&stats).Error; err != nil {
		t.Fatalf("read stats: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}

	got := stats[0]
	if got.SentPackets != 5 || got.ReceivedPackets != 4 {
		t.Fatalf("unexpected packet counters: sent=%d received=%d", got.SentPackets, got.ReceivedPackets)
	}
	if got.PacketLossPercent != 20 {
		t.Fatalf("unexpected packet loss: %.2f", got.PacketLossPercent)
	}
	if got.Platform != "esp32" {
		t.Fatalf("unexpected platform: %s", got.Platform)
	}
}

func TestSaveStatsRejectsInvalidTelemetry(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(token, IngestTelemetry{
		LatencyMs:         12.5,
		SentPackets:       2,
		ReceivedPackets:   3,
		PacketLossPercent: -1,
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if err != ErrInvalidTelemetry {
		t.Fatalf("expected ErrInvalidTelemetry, got %v", err)
	}
}

func TestSaveStatsRejectsUnknownToken(t *testing.T) {
	setupTestDB(t)
	seedDevice(t)

	err := SaveStats("unknown-token", IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       5,
		ReceivedPackets:   5,
		PacketLossPercent: 0,
	})
	if err == nil {
		t.Fatalf("expected invalid token error")
	}
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
