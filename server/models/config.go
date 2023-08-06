package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func ConnectDb() {
	database, err := gorm.Open(sqlite.Open("./storage/db/scrapper.db"), &gorm.Config{})

	if err != nil {
		log.Fatalln("Failed to connect to database!")
	}

	database.AutoMigrate(&Network{}, &Device{}, &Stat{})

	DB = database
}
