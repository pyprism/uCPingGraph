package models

import (
	"gorm.io/gorm"
	"log"
)

type Network struct {
	gorm.Model
	Name string `gorm:"unique;not null;size:500"`
}

func (n *Network) CreateNetwork(name string) (error, uint) {
	n.Name = name
	err := DB.Create(n)
	if err != nil {
		log.Println("Failed to create network! Error: ", err.Error)
		return err.Error, 0
	}
	return nil, n.ID
}

func (n *Network) GetNetworkIdByName(name string) (error, uint) {
	err := DB.Where("name = ?", name).First(n)
	if err.Error != nil {
		return err.Error, 0
	}
	return nil, n.ID
}

func (n *Network) GetAllNetwork() (error, []Network) {
	var networks []Network
	err := DB.Find(&networks)
	if err.Error != nil {
		return err.Error, nil
	}
	return nil, networks
}
