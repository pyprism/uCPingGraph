package models

import (
	"log"

	"gorm.io/gorm"
)

type Network struct {
	gorm.Model
	Name string `gorm:"unique;not null;size:500"`
}

func (n *Network) CreateNetwork(name string) (uint, error) {
	n.Name = name
	err := DB.Create(n)
	if err.Error != nil {
		log.Println("Failed to create network! Error: ", err.Error)
		return 0, err.Error
	}
	return n.ID, nil
}

func (n *Network) GetNetworkIdByName(name string) (uint, error) {
	err := DB.Where("name = ?", name).First(n)
	if err.Error != nil {
		return 0, err.Error
	}
	return n.ID, nil
}

func (n *Network) GetAllNetwork() ([]Network, error) {
	var networks []Network
	err := DB.Find(&networks)
	if err.Error != nil {
		return nil, err.Error
	}
	return networks, nil
}

func (n *Network) GetAllNetworkName() ([]string, error) {
	var networks []Network
	var names []string
	err := DB.Find(&networks)
	if err.Error != nil {
		return nil, err.Error
	}
	for _, network := range networks {
		names = append(names, network.Name)
	}
	return names, nil
}
