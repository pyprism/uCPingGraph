package models

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/gorm"
)

type Stat struct {
	gorm.Model
	NetworkID         uint
	Network           Network
	DeviceID          uint
	Device            Device
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
	Labels           []string  `json:"labels"`
	LatencySeries    []float64 `json:"latency_series"`
	PacketLossSeries []float64 `json:"packet_loss_series"`
}

type MetricsSummary struct {
	Samples            int     `json:"samples"`
	AverageLatencyMs   float64 `json:"average_latency_ms"`
	AveragePacketLoss  float64 `json:"average_packet_loss_percent"`
	Availability       float64 `json:"availability_percent"`
	LatestLatencyMs    float64 `json:"latest_latency_ms"`
	LatestPacketLoss   float64 `json:"latest_packet_loss_percent"`
	LastUpdatedRFC3339 string  `json:"last_updated"`
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

	series := EChartData{
		Labels:           make([]string, 0, len(stats)),
		LatencySeries:    make([]float64, 0, len(stats)),
		PacketLossSeries: make([]float64, 0, len(stats)),
	}

	var latencyTotal float64
	var lossTotal float64
	var sentTotal int
	var receivedTotal int

	for _, stat := range stats {
		series.Labels = append(series.Labels, stat.CreatedAt.Format("02 Jan 15:04:05"))
		series.LatencySeries = append(series.LatencySeries, stat.LatencyMs)
		series.PacketLossSeries = append(series.PacketLossSeries, stat.PacketLossPercent)
		latencyTotal += stat.LatencyMs
		lossTotal += stat.PacketLossPercent
		sentTotal += stat.SentPackets
		receivedTotal += stat.ReceivedPackets
	}

	summary := MetricsSummary{}
	if len(stats) > 0 {
		latest := stats[len(stats)-1]
		summary = MetricsSummary{
			Samples:            len(stats),
			AverageLatencyMs:   latencyTotal / float64(len(stats)),
			AveragePacketLoss:  lossTotal / float64(len(stats)),
			LatestLatencyMs:    latest.LatencyMs,
			LatestPacketLoss:   latest.PacketLossPercent,
			LastUpdatedRFC3339: latest.CreatedAt.UTC().Format(time.RFC3339),
		}
		if sentTotal > 0 {
			summary.Availability = (float64(receivedTotal) / float64(sentTotal)) * 100
		}
	}

	return &MetricsResponse{
		Series:  series,
		Summary: summary,
	}, nil
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

	xDaysAgo := time.Now().AddDate(0, 0, -days)
	result := DB.Where("created_at <= ?", xDaysAgo).Delete(&Stat{})
	return result.Error
}
