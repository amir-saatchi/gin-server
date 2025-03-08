package db

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "log"

	"github.com/amir-saatchi/rest-api/internal/models"
)

// DB struct to hold the database instance
var DB *gorm.DB

func InitDB() {
    dsn := "host=localhost user=postgres password=yourpassword dbname=mydb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    DB = db

    // Auto-migrate schema (for development only)
    AutoMigrate()
}

func AutoMigrate() {
    DB.AutoMigrate(&models.User{})
}