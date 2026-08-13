package models

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func statsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared",
		strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&Network{}, &Device{}, &Stat{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	SetDB(db)
	return db
}

func seedStatsDevice(t *testing.T) (uint, uint) {
	t.Helper()
	n := Network{}
	nid, err := n.CreateNetwork("test-net")
	if err != nil {
		t.Fatalf("create network: %v", err)
	}
	d := Device{}
	did, _, err := d.CreateDevice(int(nid), "test-device")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}
	return nid, did
}

func TestCreateStat(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	s := Stat{}
	err := s.CreateStat(int(nid), int(did), TelemetryRecord{
		LatencyMs:         10.5,
		SentPackets:       5,
		ReceivedPackets:   4,
		PacketLossPercent: 20,
		Target:            "8.8.8.8",
		Platform:          "esp32",
		RSSI:              -65,
	})
	if err != nil {
		t.Fatalf("CreateStat: %v", err)
	}

	var count int64
	DB.Model(&Stat{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 stat, got %d", count)
	}
}

func TestCreateStatNilDB(t *testing.T) {
	SetDB(nil)
	s := Stat{}
	err := s.CreateStat(1, 1, TelemetryRecord{})
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

func TestGetStatsReturnsMetrics(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	// Insert 3 records
	for i := 0; i < 3; i++ {
		s := Stat{}
		s.CreateStat(int(nid), int(did), TelemetryRecord{
			LatencyMs:         float64(10 + i),
			SentPackets:       5,
			ReceivedPackets:   5,
			PacketLossPercent: 0,
			Target:            "1.1.1.1",
			Platform:          "esp32",
			RSSI:              -50,
		})
	}

	s := Stat{}
	resp, err := s.GetStats(nid, did, 60)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if resp.Summary.Samples != 3 {
		t.Fatalf("expected 3 samples, got %d", resp.Summary.Samples)
	}
	if len(resp.Series.LatencySeries) != 3 {
		t.Fatalf("expected 3 latency points, got %d", len(resp.Series.LatencySeries))
	}
	if len(resp.Series.PacketLossSeries) != 3 {
		t.Fatalf("expected 3 loss points, got %d", len(resp.Series.PacketLossSeries))
	}
	if resp.Summary.Availability == nil || *resp.Summary.Availability != 100 {
		t.Fatalf("expected 100%% availability, got %v", resp.Summary.Availability)
	}
	// Average latency = (10+11+12)/3 = 11
	if resp.Summary.AverageLatencyMs == nil || *resp.Summary.AverageLatencyMs != 11 {
		t.Fatalf("expected average latency 11, got %v", resp.Summary.AverageLatencyMs)
	}
}

func TestGetStatsDefaultsMinute(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	s := Stat{}
	resp, err := s.GetStats(nid, did, 0) // 0 should default to 60
	if err != nil {
		t.Fatalf("GetStats with 0 minute: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetStatsEmptyResult(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	s := Stat{}
	resp, err := s.GetStats(nid, did, 60)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if resp.Summary.Samples != 0 {
		t.Fatalf("expected 0 samples, got %d", resp.Summary.Samples)
	}
	if resp.Summary.AverageLatencyMs != nil {
		t.Fatalf("expected nil average latency for empty result, got %v", *resp.Summary.AverageLatencyMs)
	}
	if resp.Summary.Availability != nil {
		t.Fatalf("expected nil availability for empty result, got %v", *resp.Summary.Availability)
	}
	if resp.Summary.LastUpdatedRFC3339 != "" {
		t.Fatalf("expected empty last_updated for empty result, got %q", resp.Summary.LastUpdatedRFC3339)
	}
}

func TestGetStatsNilDB(t *testing.T) {
	SetDB(nil)
	s := Stat{}
	_, err := s.GetStats(1, 1, 60)
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

func TestCleanupRemovesOldRecords(t *testing.T) {
	db := statsTestDB(t)
	nid, did := seedStatsDevice(t)

	// Insert an old record (40 days ago)
	oldStat := Stat{
		NetworkID:         nid,
		DeviceID:          did,
		LatencyMs:         5,
		SentPackets:       1,
		ReceivedPackets:   1,
		PacketLossPercent: 0,
		Target:            "1.1.1.1",
		Platform:          "test",
	}
	if err := db.Create(&oldStat).Error; err != nil {
		t.Fatalf("create old stat: %v", err)
	}
	// Manually set created_at to 40 days ago
	db.Exec("UPDATE stats SET created_at = ? WHERE id = ?",
		time.Now().AddDate(0, 0, -40), oldStat.ID)

	// Insert a recent record
	newStat := Stat{
		NetworkID:         nid,
		DeviceID:          did,
		LatencyMs:         10,
		SentPackets:       1,
		ReceivedPackets:   1,
		PacketLossPercent: 0,
	}
	db.Create(&newStat)

	t.Setenv("CLEANUP_DAYS", "30")

	s := Stat{}
	if err := s.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	var count int64
	db.Model(&Stat{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 stat after cleanup, got %d", count)
	}
}

func TestCleanupInvalidDays(t *testing.T) {
	statsTestDB(t)
	t.Setenv("CLEANUP_DAYS", "abc")

	s := Stat{}
	err := s.Cleanup()
	if err == nil {
		t.Fatal("expected error for invalid CLEANUP_DAYS")
	}
}

func TestCleanupRejectsNonPositiveDays(t *testing.T) {
	for _, days := range []string{"0", "-1"} {
		t.Run(days, func(t *testing.T) {
			db := statsTestDB(t)
			nid, did := seedStatsDevice(t)
			db.Create(&Stat{NetworkID: nid, DeviceID: did, SentPackets: 1, ReceivedPackets: 1})

			t.Setenv("CLEANUP_DAYS", days)

			s := Stat{}
			if err := s.Cleanup(); err == nil {
				t.Fatalf("expected error for CLEANUP_DAYS=%s", days)
			}

			var count int64
			db.Model(&Stat{}).Count(&count)
			if count != 1 {
				t.Fatalf("expected no rows deleted for CLEANUP_DAYS=%s, got count=%d", days, count)
			}
		})
	}
}

func TestCleanupNilDB(t *testing.T) {
	SetDB(nil)
	s := Stat{}
	err := s.Cleanup()
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

func TestGetStatsExcludesTotalLossFromLatencyAverage(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	s := Stat{}
	// A healthy sample.
	s.CreateStat(int(nid), int(did), TelemetryRecord{
		LatencyMs:       20,
		SentPackets:     5,
		ReceivedPackets: 5,
		Target:          "1.1.1.1",
		Platform:        "esp32",
	})
	// A total-loss probe: some firmware reports latency_ms=0 here.
	s2 := Stat{}
	s2.CreateStat(int(nid), int(did), TelemetryRecord{
		LatencyMs:         0,
		SentPackets:       5,
		ReceivedPackets:   0,
		PacketLossPercent: 100,
		Target:            "1.1.1.1",
		Platform:          "esp32",
	})

	stat := Stat{}
	resp, err := stat.GetStats(nid, did, 60)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if resp.Summary.Samples != 2 {
		t.Fatalf("expected 2 samples, got %d", resp.Summary.Samples)
	}
	// Average latency must only count the sample with a real reading.
	if resp.Summary.AverageLatencyMs == nil || *resp.Summary.AverageLatencyMs != 20 {
		t.Fatalf("expected average latency 20, got %v", resp.Summary.AverageLatencyMs)
	}
	// The latest sample was a total outage, so latest latency is unknown.
	if resp.Summary.LatestLatencyMs != nil {
		t.Fatalf("expected nil latest latency after total loss, got %v", *resp.Summary.LatestLatencyMs)
	}
	if len(resp.Series.LatencySeries) != 2 {
		t.Fatalf("expected 2 latency points, got %d", len(resp.Series.LatencySeries))
	}
	if resp.Series.LatencySeries[1] != nil {
		t.Fatalf("expected nil latency point for total-loss sample, got %v", *resp.Series.LatencySeries[1])
	}
}

func TestBuildSeriesUnderCapReturnsPerPointSeries(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	stats := []Stat{
		{CreatedAt: base, LatencyMs: 10, ReceivedPackets: 5, PacketLossPercent: 0},
		{CreatedAt: base.Add(time.Second), LatencyMs: 0, ReceivedPackets: 0, PacketLossPercent: 100},
	}

	series := buildSeries(stats)

	if len(series.Labels) != 2 || len(series.LatencySeries) != 2 || len(series.PacketLossSeries) != 2 {
		t.Fatalf("expected 2 points in every array, got labels=%d latency=%d loss=%d",
			len(series.Labels), len(series.LatencySeries), len(series.PacketLossSeries))
	}
	if series.Labels[0] != base.Format(time.RFC3339) {
		t.Fatalf("expected label %q, got %q", base.Format(time.RFC3339), series.Labels[0])
	}
	if series.LatencySeries[0] == nil || *series.LatencySeries[0] != 10 {
		t.Fatalf("expected latency point 10, got %v", series.LatencySeries[0])
	}
	if series.LatencySeries[1] != nil {
		t.Fatalf("expected nil latency point for total-loss sample, got %v", *series.LatencySeries[1])
	}
	if series.PacketLossSeries[1] != 100 {
		t.Fatalf("expected packet loss 100, got %v", series.PacketLossSeries[1])
	}
}

func TestBucketSeriesAveragesAndFlagsTotalLossBuckets(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(offset int, latency float64, received int, loss float64) Stat {
		return Stat{
			CreatedAt:         base.Add(time.Duration(offset) * time.Second),
			LatencyMs:         latency,
			ReceivedPackets:   received,
			PacketLossPercent: loss,
		}
	}

	stats := []Stat{
		// bucket 0: all received, avg latency 20, avg loss 0
		mk(0, 10, 5, 0),
		mk(1, 20, 5, 0),
		mk(2, 30, 5, 0),
		// bucket 1: one total-loss sample excluded from the latency average
		mk(3, 40, 5, 0),
		mk(4, 0, 0, 100),
		mk(5, 60, 5, 0),
		// bucket 2: single sample, total loss -> nil latency point
		mk(6, 0, 0, 100),
	}

	series := bucketSeries(stats, 3)

	if len(series.Labels) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(series.Labels))
	}

	if series.LatencySeries[0] == nil || *series.LatencySeries[0] != 20 {
		t.Fatalf("expected bucket 0 avg latency 20, got %v", series.LatencySeries[0])
	}
	if series.LatencySeries[1] == nil || *series.LatencySeries[1] != 50 {
		t.Fatalf("expected bucket 1 avg latency 50 (excluding total-loss sample), got %v", series.LatencySeries[1])
	}
	if series.LatencySeries[2] != nil {
		t.Fatalf("expected bucket 2 (all total-loss) to be nil, got %v", *series.LatencySeries[2])
	}

	wantLoss := []float64{0, float64(100) / 3, 100}
	for i, want := range wantLoss {
		if series.PacketLossSeries[i] != want {
			t.Fatalf("bucket %d: expected packet loss %.4f, got %.4f", i, want, series.PacketLossSeries[i])
		}
	}

	// Each bucket's label is taken from its last sample.
	wantLabels := []string{
		base.Add(2 * time.Second).Format(time.RFC3339),
		base.Add(5 * time.Second).Format(time.RFC3339),
		base.Add(6 * time.Second).Format(time.RFC3339),
	}
	for i, want := range wantLabels {
		if series.Labels[i] != want {
			t.Fatalf("bucket %d: expected label %q, got %q", i, want, series.Labels[i])
		}
	}
}

func TestGetStatsDownsamplesLargeResultSet(t *testing.T) {
	db := statsTestDB(t)
	nid, did := seedStatsDevice(t)

	total := maxSeriesPoints + 10
	batch := make([]Stat, 0, total)
	for i := 0; i < total; i++ {
		batch = append(batch, Stat{
			NetworkID:         nid,
			DeviceID:          did,
			LatencyMs:         float64(i % 50),
			SentPackets:       5,
			ReceivedPackets:   5,
			PacketLossPercent: 0,
			Target:            "1.1.1.1",
			Platform:          "esp32",
		})
	}
	if err := db.Create(&batch).Error; err != nil {
		t.Fatalf("batch create: %v", err)
	}

	s := Stat{}
	resp, err := s.GetStats(nid, did, 60)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if resp.Summary.Samples != total {
		t.Fatalf("expected %d samples in summary, got %d", total, resp.Summary.Samples)
	}
	// The summary must still reflect every raw row even though the series
	// gets downsampled below.
	if len(resp.Series.Labels) > maxSeriesPoints || len(resp.Series.Labels) == total {
		t.Fatalf("expected series downsampled to at most %d points (and fewer than %d raw rows), got %d",
			maxSeriesPoints, total, len(resp.Series.Labels))
	}
	if len(resp.Series.LatencySeries) != len(resp.Series.Labels) || len(resp.Series.PacketLossSeries) != len(resp.Series.Labels) {
		t.Fatalf("expected latency/loss series to match label count %d, got %d/%d",
			len(resp.Series.Labels), len(resp.Series.LatencySeries), len(resp.Series.PacketLossSeries))
	}
}

func TestGetStatsAvailabilityCalculation(t *testing.T) {
	statsTestDB(t)
	nid, did := seedStatsDevice(t)

	// Insert record with 50% packet loss
	s := Stat{}
	s.CreateStat(int(nid), int(did), TelemetryRecord{
		LatencyMs:         15,
		SentPackets:       10,
		ReceivedPackets:   5,
		PacketLossPercent: 50,
		Target:            "8.8.8.8",
		Platform:          "esp32",
		RSSI:              -70,
	})

	stat := Stat{}
	resp, err := stat.GetStats(nid, did, 60)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if resp.Summary.Availability == nil || *resp.Summary.Availability != 50 {
		t.Fatalf("expected 50%% availability, got %v", resp.Summary.Availability)
	}
}
