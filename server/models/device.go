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

type DeviceAPI struct {
	Name string `json:"name"`
}

// CheckDeviceNameIsUnique checks if the device name is unique in the network
func (d *Device) CheckDeviceNameIsUnique(networkID int, name string) bool {
	var device Device
	DB.Where("network_id = ? AND name = ?", networkID, name).First(&device)
	if device.ID > 0 {
		return false
	}
	return true
}

func (d *Device) CreateDevice(networkID int, name string) (uint, string, error) {
	token := utils.GenToken(15)
	d.NetworkID = networkID
	d.Name = name
	d.Token = token
	err := DB.Create(d)
	if err.Error != nil {
		return 0, "", err.Error
	}
	return d.ID, d.Token, nil
}

func (d *Device) GetDeviceByToken(token string) (uint, int, error) {
	err := DB.Where("token = ?", token).First(d)
	if err.Error != nil {
		return 0, 0, err.Error
	}
	return d.ID, d.NetworkID, nil
}

func (d *Device) GetDevicesByNetwork(networkID int) ([]DeviceAPI, error) {
	var devices []DeviceAPI
	err := DB.Model(&Device{}).Where("network_id = ?", networkID).Find(&devices)
	if err.Error != nil {
		return []DeviceAPI{}, err.Error
	}
	return devices, nil
}

func (d *Device) GetDeviceIdByName(name string, networkId uint) (uint, error) {
	err := DB.Where("network_id = ? AND name = ?", networkId, name).First(d)
	if err.Error != nil {
		return 0, err.Error
	}
	return d.ID, nil
}
