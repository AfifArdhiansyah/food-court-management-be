package main

import (
	"log"

	"foodcourt-backend/internal/config"
	"foodcourt-backend/internal/database"
)

func main() {
	log.Println("Starting database seeding...")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run seeder
	if err := db.Seed(); err != nil {
		log.Fatalf("Failed to run seeder: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}
