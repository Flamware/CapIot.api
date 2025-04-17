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

// GetLocationByID retrieves a location by its ID from the database.
func (r *PostgresLocationRepository) GetLocationByID(id string) (models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT location_id, location_name, location_description FROM locations WHERE location_id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var location models.Location
	if err := row.Scan(&location.ID, &location.Name, &location.Description); err != nil {
		if err == sql.ErrNoRows {
			return models.Location{}, nil // Return empty location if not found
		}
		log.Printf("Error scanning location row: %v\n", err)
		return models.Location{}, err
	}

	return location, nil
}

// GetCaptorsByLocationID retrieves captors associated with a location ID.
func (r *PostgresLocationRepository) GetCaptorsByLocationID(locationID string) ([]models.Captor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT
       c.captor_id,
       c.captor_type
    FROM captors c
    JOIN device_captors dc ON c.captor_id = dc.captor_id
    JOIN device_location dl ON dc.device_id = dl.device_id
    WHERE dl.location_id = $1 `

	rows, err := r.db.QueryContext(ctx, query, locationID)
	if err != nil {
		log.Printf("Error getting captors by location ID: %v\n", err)
		return []models.Captor{}, err
	}
	defer rows.Close()

	captors := []models.Captor{} // Initialize as an empty slice

	for rows.Next() {
		var captor models.Captor

		if err := rows.Scan(
			&captor.CaptorID,
			&captor.CaptorType,
		); err != nil {
			log.Printf("Error scanning captor row: %v\n", err)
			return []models.Captor{}, err
		}
		captors = append(captors, captor)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating captor rows: %v\n", err)
		return []models.Captor{}, err
	}

	return captors, nil
}

// GetDevicesByLocationID retrieves devices associated with a location ID.
func (r *PostgresLocationRepository) GetDevicesByLocationID(id string) ([]models.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT d.device_id, d.status, d.last_seen
	FROM devices d
	JOIN device_location dl ON d.device_id = dl.device_id
	WHERE dl.location_id = $1`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		log.Printf("Error getting devices by location ID: %v\n", err)
		return []models.Device{}, err
	}
	defer rows.Close()

	devices := []models.Device{}

	for rows.Next() {
		var device models.Device

		if err := rows.Scan(
			&device.DeviceID,
			&device.Status,
			&device.LastSeen,
		); err != nil {
			log.Printf("Error scanning device row: %v\n", err)
			return []models.Device{}, err
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating device rows: %v\n", err)
		return []models.Device{}, err
	}

	return devices, nil
}
