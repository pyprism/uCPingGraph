package models

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pyprism/uCPingGraph/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDb() error {
	path := utils.GetEnv("DB_PATH", "./storage/db/uCPingGraph.db")
	return ConnectDbWithPath(path)
}

func ConnectDbWithPath(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create db directory: %w", err)
	}

	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	if err := database.AutoMigrate(&Network{}, &Device{}, &Stat{}); err != nil {
		return fmt.Errorf("auto-migrate db: %w", err)
	}

	DB = database
	return nil
}

func SetDB(db *gorm.DB) {
	DB = db
}
