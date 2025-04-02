package repository

import (
	"api.cap.iot/dao"
	"api.cap.iot/models"
	"context"
	"database/sql"
	"log"
	"time"
)

// PostgresDeviceRepository implements DeviceDAO using PostgreSQL.
type PostgresDeviceRepository struct {
	db *sql.DB
}

// NewPostgresDeviceRepository creates a new PostgresDeviceRepository.
func NewPostgresDeviceRepository(db *sql.DB) dao.DeviceDAO {
	return &PostgresDeviceRepository{db: db}
}

// InsertDevice inserts a device into the database.
func (r *PostgresDeviceRepository) InsertDevice(device models.Device) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO devices (device_id, timestamp) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, device.DeviceID, device.Timestamp)
	if err != nil {
		log.Printf("Error inserting device into database: %v\n", err)
		return err
	}
	log.Printf("Device %s inserted into database.\n", device.DeviceID)
	return nil
}

// DeviceExists checks if a device exists in the database.
func (r *PostgresDeviceRepository) DeviceExists(deviceID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if device exists: %v\n", err)
		return false, err
	}
	return exists, nil
}

// GetAllDevices retrieves all devices from the database.
func (r *PostgresDeviceRepository) GetAllDevices() ([]models.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT device_id, timestamp FROM devices`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error getting all devices: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(&device.DeviceID, &device.Timestamp); err != nil {
			log.Printf("Error scanning device row: %v\n", err)
			return nil, err
		}
		devices = append(devices, device)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating device rows: %v\n", err)
		return nil, err
	}

	return devices, nil
}

// IsDeviceAssigned checks if a device is already assigned to a location.
func (r *PostgresDeviceRepository) IsDeviceAssigned(deviceID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT EXISTS(SELECT 1 FROM device_location WHERE device_id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking if device is assigned: %v\n", err)
		return false, err
	}
	return exists, nil
}

// SetDeviceToLocation associates a device with a location in the database.
func (r *PostgresDeviceRepository) SetDeviceToLocation(deviceID string, locationID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare the query to insert the device-location assignment
	query := `INSERT INTO device_location (device_id, location_id, assigned_at) 
              VALUES ($1, $2, $3)`

	// Execute the query
	_, err := r.db.ExecContext(ctx, query, deviceID, locationID, time.Now())
	if err != nil {
		log.Printf("Error associating device %s with location %d: %v\n", deviceID, locationID, err)
		return err
	}

	log.Printf("Device %s associated with location %d.\n", deviceID, locationID)
	return nil
}
