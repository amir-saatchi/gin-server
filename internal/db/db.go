package db

import (
    "log"
)

var DBS_Manager *DBManager // Global variable to hold the DBManager instance

func InitDB() {
    // Load database configurations from environment variables or a config file
    dsnMap := map[string]string{
        "company1": "postgresql://user1:password1@postgres-company1:5432/company1db",
        "company2": "postgresql://user2:password2@postgres-company2:5432/company2db",
        // Add more companies as needed
    }

    // Expose the manager via a getter function
    DBS_Manager = NewDBManager(dsnMap)
    log.Println("Database manager initialized")
}