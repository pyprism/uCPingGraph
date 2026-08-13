package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pyprism/uCPingGraph/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSeriesTestDB(t *testing.T) string {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared",
		strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Network{}, &models.Device{}, &models.Stat{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	models.SetDB(db)

	network := models.Network{}
	networkID, err := network.CreateNetwork("office")
	if err != nil {
		t.Fatalf("create network: %v", err)
	}

	device := models.Device{}
	_, token, err := device.CreateDevice(int(networkID), "sensor-1")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}
	return token
}

func TestGetSeriesReturnsData(t *testing.T) {
	token := setupSeriesTestDB(t)

	// Seed a telemetry record
	if err := SaveStats(token, IngestTelemetry{
		LatencyMs:         25.0,
		SentPackets:       10,
		ReceivedPackets:   9,
		PacketLossPercent: 10,
		Target:            "8.8.8.8",
		Platform:          "esp32",
		RSSI:              -55,
	}); err != nil {
		t.Fatalf("SaveStats: %v", err)
	}

	resp, err := GetSeries("office", "sensor-1", 60)
	if err != nil {
		t.Fatalf("GetSeries: %v", err)
	}
	if resp.Summary.Samples != 1 {
		t.Fatalf("expected 1 sample, got %d", resp.Summary.Samples)
	}
	if resp.Series.LatencySeries[0] == nil || *resp.Series.LatencySeries[0] != 25.0 {
		t.Fatalf("expected latency 25.0, got %v", resp.Series.LatencySeries[0])
	}
}

func TestGetSeriesNetworkNotFound(t *testing.T) {
	setupSeriesTestDB(t)

	_, err := GetSeries("nonexistent", "sensor-1", 60)
	if err == nil {
		t.Fatal("expected error for nonexistent network")
	}
}

func TestGetSeriesDeviceNotFound(t *testing.T) {
	setupSeriesTestDB(t)

	_, err := GetSeries("office", "nonexistent", 60)
	if err == nil {
		t.Fatal("expected error for nonexistent device")
	}
}

func TestSaveStatsEmptyToken(t *testing.T) {
	setupSeriesTestDB(t)

	err := SaveStats("", IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       1,
		ReceivedPackets:   1,
		PacketLossPercent: 0,
	})
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestSaveStatsWhitespaceOnlyToken(t *testing.T) {
	setupSeriesTestDB(t)

	err := SaveStats("   ", IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       1,
		ReceivedPackets:   1,
		PacketLossPercent: 0,
	})
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestSaveStatsPacketLossOver100(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(token, IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       5,
		ReceivedPackets:   5,
		PacketLossPercent: 150,
	})
	if err != ErrInvalidTelemetry {
		t.Fatalf("expected ErrInvalidTelemetry for loss > 100, got %v", err)
	}
}

func TestSaveStatsReceivedMoreThanSent(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(token, IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       3,
		ReceivedPackets:   5,
		PacketLossPercent: 0,
	})
	if err != ErrInvalidTelemetry {
		t.Fatalf("expected ErrInvalidTelemetry for received > sent, got %v", err)
	}
}

func TestSaveStatsNegativeSentPackets(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(token, IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       -1,
		ReceivedPackets:   0,
		PacketLossPercent: 0,
	})
	if err != ErrInvalidTelemetry {
		t.Fatalf("expected ErrInvalidTelemetry for negative sent, got %v", err)
	}
}

func TestSaveStatsTrimsWhitespace(t *testing.T) {
	setupTestDB(t)
	token := seedDevice(t)

	err := SaveStats(" "+token+" ", IngestTelemetry{
		LatencyMs:         10,
		SentPackets:       5,
		ReceivedPackets:   5,
		PacketLossPercent: 0,
		Target:            "  1.1.1.1  ",
		Platform:          "  esp32  ",
		RSSI:              -60,
	})
	if err != nil {
		t.Fatalf("SaveStats with whitespace: %v", err)
	}

	// Verify trimmed values
	var stat models.Stat
	models.DB.Last(&stat)
	if stat.Target != "1.1.1.1" {
		t.Fatalf("expected trimmed target '1.1.1.1', got %q", stat.Target)
	}
	if stat.Platform != "esp32" {
		t.Fatalf("expected trimmed platform 'esp32', got %q", stat.Platform)
	}
}
