package models

import "gorm.io/gorm"

type Device struct {
	gorm.Model
	Network Network `gorm:"foreignKey:NetworkID"`
	Name    string  `gorm:"not null;size:500"`
	Token   string  `gorm:"not null;size:15"`
}
