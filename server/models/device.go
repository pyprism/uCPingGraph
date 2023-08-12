package models

import (
	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/gorm"
)

type Device struct {
	gorm.Model
	NetworkID int
	Network   Network
	Name      string `gorm:"not null;size:500"`
	Token     string `gorm:"index;not null;size:15"`
}

func (d *Device) CreateDevice(networkID int, name string) (error, uint, string) {
	token := utils.GenToken(15)
	d.NetworkID = networkID
	d.Name = name
	d.Token = token
	err := DB.Create(d)
	if err != nil {
		return err.Error, 0, ""
	}
	return nil, d.ID, d.Token
}
