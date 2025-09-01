package config

import (
	"fmt"
	"log"
	"os"

	// "ssb_api/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database!", err)
	}

	// Set dulu DB sebelum digunakan
	DB = database

	// Sekarang AutoMigrate aman
	err = DB.AutoMigrate(
	// &models.User{},
	// &models.Event{},
	// &models.Vendor{},
	// &models.Challenge{},
	// &models.ChallengeLog{},
	// &models.Payment{},
	// &models.Training{},
	// &models.Match{},
	// &models.EventLog{},
	)
	if err != nil {
		log.Fatal("❌ AutoMigrate failed: ", err)
	}

	fmt.Println("✅ Database connected and migrated!")
}
