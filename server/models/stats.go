package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/gorm"
)

// maxSeriesPoints caps how many points /api/series returns. Windows with more
// raw samples than this are downsampled into evenly-sized, averaged buckets
// so a 10080-minute request can't ship ~120k points per array to the browser.
const maxSeriesPoints = 720

type Stat struct {
	gorm.Model
	NetworkID uint `gorm:"index:idx_stats_lookup,priority:1"`
	Network   Network
	DeviceID  uint `gorm:"index:idx_stats_lookup,priority:2"`
	Device    Device
	// CreatedAt shadows gorm.Model's field so the composite index below can
	// be declared on it; the type and semantics are unchanged.
	CreatedAt         time.Time `gorm:"index:idx_stats_lookup,priority:3"`
	LatencyMs         float64
	SentPackets       int
	ReceivedPackets   int
	PacketLossPercent float64
	Target            string `gorm:"size:64"`
	Platform          string `gorm:"size:32"`
	RSSI              int
}

type TelemetryRecord struct {
	LatencyMs         float64
	SentPackets       int
	ReceivedPackets   int
	PacketLossPercent float64
	Target            string
	Platform          string
	RSSI              int
}

type EChartData struct {
	Labels           []string   `json:"labels"`
	LatencySeries    []*float64 `json:"latency_series"`
	PacketLossSeries []float64  `json:"packet_loss_series"`
}

type MetricsSummary struct {
	Samples            int      `json:"samples"`
	AverageLatencyMs   *float64 `json:"average_latency_ms"`
	AveragePacketLoss  *float64 `json:"average_packet_loss_percent"`
	Availability       *float64 `json:"availability_percent"`
	LatestLatencyMs    *float64 `json:"latest_latency_ms"`
	LatestPacketLoss   *float64 `json:"latest_packet_loss_percent"`
	LastUpdatedRFC3339 string   `json:"last_updated"`
}

type MetricsResponse struct {
	Series  EChartData     `json:"series"`
	Summary MetricsSummary `json:"summary"`
}

func (s *Stat) CreateStat(networkID int, deviceID int, record TelemetryRecord) error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	s.NetworkID = uint(networkID)
	s.DeviceID = uint(deviceID)
	s.LatencyMs = record.LatencyMs
	s.SentPackets = record.SentPackets
	s.ReceivedPackets = record.ReceivedPackets
	s.PacketLossPercent = record.PacketLossPercent
	s.Target = record.Target
	s.Platform = record.Platform
	s.RSSI = record.RSSI

	result := DB.Create(s)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *Stat) GetStats(networkID, deviceID uint, minute int) (*MetricsResponse, error) {
	if DB == nil {
		return nil, errors.New("database is not initialized")
	}

	if minute <= 0 {
		// Defensive default only; the sole caller (controllers.Series) already
		// validates 1 <= minute <= 10080 before this is reached.
		minute = 60
	}

	xMinAgo := time.Now().Add(-time.Duration(minute) * time.Minute)
	var stats []Stat
	result := DB.
		Where("created_at >= ? AND network_id = ? AND device_id = ?", xMinAgo, networkID, deviceID).
		Order("created_at ASC").
		Find(&stats)
	if result.Error != nil {
		return nil, result.Error
	}

	summary := MetricsSummary{}
	if len(stats) > 0 {
		var latencyTotal float64
		var latencySamples int
		var lossTotal float64
		var sentTotal int
		var receivedTotal int

		for _, stat := range stats {
			lossTotal += stat.PacketLossPercent
			sentTotal += stat.SentPackets
			receivedTotal += stat.ReceivedPackets
			// A fully-lost probe (ReceivedPackets == 0) carries no real
			// latency measurement; excluding it keeps an outage from
			// dragging the average latency toward zero.
			if stat.ReceivedPackets > 0 {
				latencyTotal += stat.LatencyMs
				latencySamples++
			}
		}

		latest := stats[len(stats)-1]
		latestLoss := latest.PacketLossPercent
		avgLoss := lossTotal / float64(len(stats))
		summary = MetricsSummary{
			Samples:            len(stats),
			AveragePacketLoss:  &avgLoss,
			LatestPacketLoss:   &latestLoss,
			LastUpdatedRFC3339: latest.CreatedAt.UTC().Format(time.RFC3339),
		}
		if latencySamples > 0 {
			avgLatency := latencyTotal / float64(latencySamples)
			summary.AverageLatencyMs = &avgLatency
		}
		if latest.ReceivedPackets > 0 {
			latestLatency := latest.LatencyMs
			summary.LatestLatencyMs = &latestLatency
		}
		if sentTotal > 0 {
			availability := (float64(receivedTotal) / float64(sentTotal)) * 100
			summary.Availability = &availability
		}
	}

	return &MetricsResponse{
		Series:  buildSeries(stats),
		Summary: summary,
	}, nil
}

func buildSeries(stats []Stat) EChartData {
	if len(stats) <= maxSeriesPoints {
		series := EChartData{
			Labels:           make([]string, 0, len(stats)),
			LatencySeries:    make([]*float64, 0, len(stats)),
			PacketLossSeries: make([]float64, 0, len(stats)),
		}
		for _, stat := range stats {
			series.Labels = append(series.Labels, stat.CreatedAt.UTC().Format(time.RFC3339))
			series.LatencySeries = append(series.LatencySeries, latencyPoint(stat))
			series.PacketLossSeries = append(series.PacketLossSeries, stat.PacketLossPercent)
		}
		return series
	}

	return bucketSeries(stats, maxSeriesPoints)
}

func latencyPoint(stat Stat) *float64 {
	if stat.ReceivedPackets == 0 {
		return nil
	}
	latency := stat.LatencyMs
	return &latency
}

// bucketSeries downsamples stats into at most n evenly-sized buckets,
// averaging latency (over samples with a real reading) and packet loss
// within each bucket.
func bucketSeries(stats []Stat, n int) EChartData {
	series := EChartData{
		Labels:           make([]string, 0, n),
		LatencySeries:    make([]*float64, 0, n),
		PacketLossSeries: make([]float64, 0, n),
	}

	total := len(stats)
	bucketSize := (total + n - 1) / n

	for start := 0; start < total; start += bucketSize {
		end := start + bucketSize
		if end > total {
			end = total
		}
		bucket := stats[start:end]

		var latencyTotal float64
		var latencySamples int
		var lossTotal float64
		for _, stat := range bucket {
			lossTotal += stat.PacketLossPercent
			if stat.ReceivedPackets > 0 {
				latencyTotal += stat.LatencyMs
				latencySamples++
			}
		}

		last := bucket[len(bucket)-1]
		series.Labels = append(series.Labels, last.CreatedAt.UTC().Format(time.RFC3339))
		series.PacketLossSeries = append(series.PacketLossSeries, lossTotal/float64(len(bucket)))
		if latencySamples > 0 {
			avgLatency := latencyTotal / float64(latencySamples)
			series.LatencySeries = append(series.LatencySeries, &avgLatency)
		} else {
			series.LatencySeries = append(series.LatencySeries, nil)
		}
	}

	return series
}

func (s *Stat) Cleanup() error {
	if DB == nil {
		return errors.New("database is not initialized")
	}

	daysValue := utils.GetEnv("CLEANUP_DAYS", "30")
	days, err := strconv.Atoi(daysValue)
	if err != nil {
		return fmt.Errorf("invalid CLEANUP_DAYS value %q: %w", daysValue, err)
	}
	if days < 1 {
		return fmt.Errorf("CLEANUP_DAYS must be >= 1, got %d", days)
	}

	xDaysAgo := time.Now().AddDate(0, 0, -days)
	result := DB.Unscoped().Where("created_at <= ?", xDaysAgo).Delete(&Stat{})
	return result.Error
}
