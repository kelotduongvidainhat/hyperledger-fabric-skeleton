package db

import (
	"backend/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(host, port, user, password, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established")

	// Auto-migrate the schemas
	err = db.AutoMigrate(&models.User{}, &models.Asset{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate: %v", err)
	}

	log.Println("Database migration completed")
	return db, nil
}
