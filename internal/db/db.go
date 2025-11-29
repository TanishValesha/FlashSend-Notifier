// docker run --name notifly-postgres \
//   -e POSTGRES_USER=postgres \
//   -e POSTGRES_PASSWORD=postgres \
//   -e POSTGRES_DB=notifly \
//   -p 5432:5432 \
//   -d postgres:latest

package db

import (
	"log"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/TanishValesha/FlashSend-Notifier/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error

	DB, err = gorm.Open(postgres.Open(config.Cfg.DB))

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to DB")
}

func CreateEnums() {
	enumSQL := `
	DO $$ BEGIN
		CREATE TYPE channel_type AS ENUM ('email', 'sms', 'whatsapp');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	DO $$ BEGIN
		CREATE TYPE status_type AS ENUM ('sent', 'failed');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;
	`

	if err := DB.Exec(enumSQL).Error; err != nil {
		log.Fatalf("Failed to create enums: %v", err)
	}
}

func AutoMigrate() {
	err := DB.AutoMigrate(&models.User{}, &models.APIKey{}, &models.Notification{})

	if err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}

	log.Println("Database AutoMigrated")
}
