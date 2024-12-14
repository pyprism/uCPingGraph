package models

import (
	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/gorm"
	"log"
	"strconv"
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

func (s *Stat) GetStats(networkID, deviceID uint, minute int) (*EChartData, error) {
	// Calculate the X min ago
	xMinAgo := time.Now().Add(-time.Duration(minute) * time.Minute)

	var stats []Stat
	if result := DB.Where("created_at >= ? AND network_id = ? AND device_id = ?", xMinAgo, networkID, deviceID).Order("created_at ASC").Find(&stats); result.Error != nil {
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

// Cleanup deletes the stats older than X days
func (s *Stat) Cleanup() error {
	days := utils.GetEnv("CLEANUP_DAYS", "30")
	daysInt, err := strconv.Atoi(days)
	if err != nil {
		log.Fatal(err)
	}
	xDaysAgo := time.Now().AddDate(0, 0, -daysInt)

	if result := DB.Where("created_at <= ?", xDaysAgo).Delete(&Stat{}); result.Error != nil {
		return result.Error
	}

	return nil
}
