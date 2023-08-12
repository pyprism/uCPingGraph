package models

import "gorm.io/gorm"

// Stat is the model for the ping stats table.
type Stat struct {
	gorm.Model
	NetworkID int
	Network   Network
	DeviceID  int
	Device    Device `gorm:"foreignKey:DeviceID;references:ID"`
	Latency   int    `gorm:"not null"`
}
