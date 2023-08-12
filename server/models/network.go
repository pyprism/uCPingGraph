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
