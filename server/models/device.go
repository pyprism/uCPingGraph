package models

import "gorm.io/gorm"

type Device struct {
	gorm.Model
	NetworkID int
	Network   Network
	Name      string `gorm:"not null;size:500"`
	Token     string `gorm:"not null;size:15"`
}
