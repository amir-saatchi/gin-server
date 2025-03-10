package main

import (
	"fmt"
	"log"

	"github.com/amir-saatchi/rest-api/internal/db"
	"github.com/amir-saatchi/rest-api/internal/routes"
)

func main() {
	// Initialize the database
	db.InitDB()
		
    // Initialize Gin router
	router := routes.NewRouter(db.MainDB,db.SecondaryDB)

    port := 8080
    addr := fmt.Sprintf(":%d", port)
    fmt.Printf("Server is running on http://localhost:%d\n", port)

    // Start the server with Gin
    err := router.Run(addr)
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}