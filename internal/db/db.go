package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/amir-saatchi/rest-api/internal/models"
)

// MainDB and SecondaryDB are the two database instances
var MainDB *gorm.DB
var SecondaryDB *gorm.DB

func InitDB() {
    // Connect to MainDB
       dsnMain := os.Getenv("DB_MAIN_URL")
       dbMain, err := gorm.Open(postgres.Open(dsnMain), &gorm.Config{})
       if err != nil {
           log.Fatalf("Failed to connect to main database: %v", err)
       }
       MainDB = dbMain
       AutoMigrateMain()

    // Connect to SecondaryDB
    dsnSecondary := os.Getenv("DB_SECONDARY_URL")
    dbSecondary, err := gorm.Open(postgres.Open(dsnSecondary), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to secondary database: %v", err)
    }
    SecondaryDB = dbSecondary
    AutoMigrateSecondary()
}

func AutoMigrateMain() {
    MainDB.AutoMigrate(&models.User{}) // Example model for MainDB
}

func AutoMigrateSecondary() {
    SecondaryDB.AutoMigrate(&models.Log{}) // Example model for SecondaryDB
}