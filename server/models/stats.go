package models

import "gorm.io/gorm"

// Stat is the model for the ping stats table.
type Stat struct {
	gorm.Model
	Network Network `gorm:"foreignKey:NetworkID"`
	Device  Device  `gorm:"foreignKey:DeviceID"`
	Latency int     `gorm:"not null"`
}
