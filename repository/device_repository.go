package repository

import (
	"api.cap.iot/models"
	"context"
	"database/sql"
	"log"
	"time"
)

type DeviceRepository interface {
	InsertDevice(device models.Device) error
	DeviceExists(deviceID string) (bool, error)
}

type PostgresDeviceRepository struct {
	db *sql.DB
}

func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

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
