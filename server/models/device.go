package models

import (
	"errors"

	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/gorm"
)

type Device struct {
	gorm.Model
	NetworkID uint `gorm:"index:idx_network_device_name,unique"`
	Network   Network
	Name      string `gorm:"not null;size:500;index:idx_network_device_name,unique"`
	Token     string `gorm:"index;unique;not null;size:64"`
}

type DeviceAPI struct {
	Name string `json:"name"`
}

// CheckDeviceNameIsUnique checks if the device name is unique in the network
func (d *Device) CheckDeviceNameIsUnique(networkID int, name string) bool {
	if DB == nil {
		return false
	}

	var device Device
	result := DB.Where("network_id = ? AND name = ?", networkID, name).First(&device)
	return errors.Is(result.Error, gorm.ErrRecordNotFound)
}

func (d *Device) CreateDevice(networkID int, name string) (uint, string, error) {
	if DB == nil {
		return 0, "", errors.New("database is not initialized")
	}

	token := utils.GenToken(32)
	d.NetworkID = uint(networkID)
	d.Name = name
	d.Token = token
	result := DB.Create(d)
	if result.Error != nil {
		return 0, "", result.Error
	}
	return d.ID, d.Token, nil
}

func (d *Device) GetDeviceByToken(token string) (uint, int, error) {
	if DB == nil {
		return 0, 0, errors.New("database is not initialized")
	}

	result := DB.Where("token = ?", token).First(d)
	if result.Error != nil {
		return 0, 0, result.Error
	}
	return d.ID, int(d.NetworkID), nil
}

func (d *Device) GetDevicesByNetwork(networkID int) ([]DeviceAPI, error) {
	if DB == nil {
		return []DeviceAPI{}, errors.New("database is not initialized")
	}

	var devices []DeviceAPI
	result := DB.Model(&Device{}).Where("network_id = ?", networkID).Order("name ASC").Find(&devices)
	if result.Error != nil {
		return []DeviceAPI{}, result.Error
	}
	return devices, nil
}

func (d *Device) GetDeviceIdByName(name string, networkId uint) (uint, error) {
	if DB == nil {
		return 0, errors.New("database is not initialized")
	}

	result := DB.Where("network_id = ? AND name = ?", networkId, name).First(d)
	if result.Error != nil {
		return 0, result.Error
	}
	return d.ID, nil
}
