package main

import (
	"log"

	"foodcourt-backend/internal/config"
	"foodcourt-backend/internal/database"
)

func main() {
	log.Println("Starting database migration...")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Database migration completed successfully!")
}
