package models

import "gorm.io/gorm"

type Network struct {
	gorm.Model
	Name string `gorm:"index:unique;not null;size:500"`
}
