package models

import (
	"errors"

	"gorm.io/gorm"
)

type Network struct {
	gorm.Model
	Name string `gorm:"unique;not null;size:500"`
}

func (n *Network) CreateNetwork(name string) (uint, error) {
	if DB == nil {
		return 0, errors.New("database is not initialized")
	}

	n.Name = name
	result := DB.Create(n)
	if result.Error != nil {
		return 0, result.Error
	}
	return n.ID, nil
}

func (n *Network) GetNetworkIdByName(name string) (uint, error) {
	if DB == nil {
		return 0, errors.New("database is not initialized")
	}

	result := DB.Where("name = ?", name).First(n)
	if result.Error != nil {
		return 0, result.Error
	}
	return n.ID, nil
}

func (n *Network) GetAllNetwork() ([]Network, error) {
	if DB == nil {
		return nil, errors.New("database is not initialized")
	}

	var networks []Network
	result := DB.Order("name ASC").Find(&networks)
	if result.Error != nil {
		return nil, result.Error
	}
	return networks, nil
}

func (n *Network) GetAllNetworkName() ([]string, error) {
	networks, err := n.GetAllNetwork()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(networks))
	for _, network := range networks {
		names = append(names, network.Name)
	}

	return names, nil
}
