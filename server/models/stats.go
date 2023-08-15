package models

import "gorm.io/gorm"

// Stat is the model for the ping stats table.
type Stat struct {
	gorm.Model
	NetworkID int
	Network   Network
	DeviceID  int
	Device    Device
	Latency   float32 `gorm:"not null"`
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
