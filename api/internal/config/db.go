package config

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // Import the PostgreSQL driver
	"log"
	"os"
)

func InitDB() (*sql.DB, error) {
	// Build the connection string from environment variables
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
		return nil, err
	}

	log.Println("✅ Successfully connected to the database")
	return db, nil
}
