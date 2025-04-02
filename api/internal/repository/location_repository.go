package repository

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"log"
	"time"
)

// PostgresLocationRepository implements LocationDAO using PostgreSQL.
type PostgresLocationRepository struct {
	db *sql.DB
}

// NewPostgresLocationRepository creates a new PostgresLocationRepository.
func NewPostgresLocationRepository(db *sql.DB) dao.LocationDAO {
	return &PostgresLocationRepository{db: db}
}

// InsertLocation inserts a location into the database.
func (r *PostgresLocationRepository) InsertLocation(location models.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO locations (location_name, location_description) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, location.Name, location.Description)
	if err != nil {
		log.Printf("Error inserting location into database: %v\n", err)
		return err
	}
	log.Printf("Location %s inserted into database.\n", location.ID)
	return nil
}

// LocationExists checks if a location exists in the database.
func (r *PostgresLocationRepository) LocationExists(ID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // Reduce timeout if 5s is too long
	defer cancel()

	const query = `SELECT 1 FROM locations WHERE location_id = $1 LIMIT 1` // Use LIMIT 1 instead of EXISTS
	var exists int
	err := r.db.QueryRowContext(ctx, query, ID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil // Return false if no rows are found
	} else if err != nil {
		log.Printf("Error checking location existence: %v\n", err)
		return false, err
	}
	return true, nil
}

// GetAllLocations retrieves all locations from the database.
func (r *PostgresLocationRepository) GetAllLocations() ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT location_id, location_name, location_description FROM locations`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error getting all locations: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var locations []models.Location
	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Description); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return nil, err
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location rows: %v\n", err)
		return nil, err
	}

	return locations, nil
}
