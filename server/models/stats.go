package models

import (
	"gorm.io/gorm"
	"time"
)

// Stat is the model for the ping stats table.
type Stat struct {
	gorm.Model
	NetworkID int
	Network   Network
	DeviceID  int
	Device    Device
	Latency   float32 `gorm:"not null"`
}

type EChartData struct {
	Labels []string  `json:"labels"`
	Series []float32 `json:"series"`
}

// CreateStat creates a new stat.
func (s *Stat) CreateStat(networkID int, deviceID int, latency float32) error {
	s.NetworkID = networkID
	s.DeviceID = deviceID
	s.Latency = latency
	err := DB.Create(s)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func (s *Stat) GetStats(networkID, deviceID uint) (*EChartData, error) {
	// Calculate the time 2 hours ago
	twoHoursAgo := time.Now().Add(-2 * time.Hour)

	var stats []Stat
	if result := DB.Where("created_at >= ? AND network_id = ? AND device_id = ?", twoHoursAgo, networkID, deviceID).Order("created_at ASC").Find(&stats); result.Error != nil {
		return nil, result.Error
	}

	chartData := EChartData{
		Labels: make([]string, len(stats)),
		Series: make([]float32, len(stats)),
	}

	for i, stat := range stats {
		chartData.Labels[i] = stat.CreatedAt.Format("02-January-2006 03:04:05 PM")
		chartData.Series[i] = stat.Latency
	}

	return &chartData, nil
}
