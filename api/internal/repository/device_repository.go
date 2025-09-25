package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// PostgresDeviceDAO implements the DeviceDAO interface using PostgreSQL.
type PostgresDeviceDAO struct {
	db *sql.DB
}

// NewPostgresDeviceDAO creates a new PostgresDeviceDAO instance.
func NewPostgresDeviceDAO(db *sql.DB) *PostgresDeviceDAO {
	return &PostgresDeviceDAO{db: db}
}

// getExecutor returns either the transaction or the database connection for execution.
func (d *PostgresDeviceDAO) getExecutor(tx *sql.Tx) interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
} {
	if tx != nil {
		return tx
	}
	return d.db
}

// BeginTransaction starts a new database transaction.
func (d *PostgresDeviceDAO) BeginTransaction() (*sql.Tx, error) {
	return d.db.Begin()
}

// CreateDevice creates a new device record within a transaction.
func (d *PostgresDeviceDAO) CreateDevice(tx *sql.Tx, device *models.Device) error {
	executor := d.getExecutor(tx)
	query := `
		INSERT INTO public.devices (device_id, last_seen, status, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (device_id) DO NOTHING`
	_, err := executor.Exec(query, device.DeviceID, device.LastSeen, device.Status, device.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}
	return nil
}

// DeviceExists checks if a device with the given ID exists.
func (d *PostgresDeviceDAO) DeviceExists(deviceID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM public.devices WHERE device_id = $1)"
	err := d.db.QueryRow(query, deviceID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if device exists: %w", err)
	}
	return exists, nil
}

// GetDeviceByID retrieves a device record by its ID.
func (d *PostgresDeviceDAO) GetDeviceByID(id string) (*models.Device, error) {
	row := d.db.QueryRow("SELECT device_id, last_seen, status, voltage, current, power, created_at FROM devices WHERE device_id = $1", id)
	var device models.Device
	err := row.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.Voltage, &device.Current, &device.Power, &device.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error scanning device with ID %s: %w", id, err)
	}
	return &device, nil
}

// UpdateDeviceLastSeenAndStatus updates the last seen timestamp and status of a device within a transaction.
func (d *PostgresDeviceDAO) UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE devices SET status = $1, last_seen = NOW() WHERE device_id = $2", status, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device last seen and status: %w", err)
	}
	return nil
}

// UpdateDeviceOperationalStatus updates the operational status of a device within a transaction.
func (d *PostgresDeviceDAO) UpdateDeviceOperationalStatus(tx *sql.Tx, id string, status models.OperationalStatus) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE devices SET status = $1 WHERE device_id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update device operational status: %w", err)
	}
	return nil
}

// GetAllDevices retrieves all device records from the database.
func (d *PostgresDeviceDAO) GetAllDevices() ([]*models.Device, error) {
	rows, err := d.db.Query("SELECT device_id, last_seen, status, voltage, current, power, created_at FROM devices")
	if err != nil {
		return nil, fmt.Errorf("failed to query all devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var dev models.Device
		if err := rows.Scan(&dev.DeviceID, &dev.LastSeen, &dev.Status, &dev.Voltage, &dev.Current, &dev.Power, &dev.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		devices = append(devices, &dev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during device rows iteration: %w", err)
	}
	return devices, nil
}

// GetUnassignedDevices retrieves devices that are not assigned to any location.
func (d *PostgresDeviceDAO) GetUnassignedDevices() ([]*models.Device, error) {
	rows, err := d.db.Query(`
		SELECT d.device_id, d.last_seen, d.status, d.created_at
		FROM devices d
		LEFT JOIN device_location dl ON d.device_id = dl.device_id AND dl.is_current = true
		WHERE dl.device_id IS NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query unassigned devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan unassigned device row: %w", err)
		}
		devices = append(devices, &device)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during unassigned device rows iteration: %w", err)
	}
	return devices, nil
}

// DeleteDevice deletes a device by its ID within a transaction.
func (d *PostgresDeviceDAO) DeleteDevice(tx *sql.Tx, id string) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("DELETE FROM devices WHERE device_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete device %s: %w", id, err)
	}
	return nil
}

// UnassignDeviceFromLocation unassigns a device from its current location within a transaction.
func (d *PostgresDeviceDAO) UnassignDeviceFromLocation(tx *sql.Tx, id string) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE device_location SET is_current = false WHERE device_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to unassign device %s from location: %w", id, err)
	}
	return nil
}

// SetDeviceToLocation assigns a device to a location within a transaction.
func (d *PostgresDeviceDAO) SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error {
	// Start a transaction.
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer a rollback in case of an error.
	defer tx.Rollback()

	// --- Use a lock to prevent a race condition ---
	// A SELECT FOR UPDATE locks the row(s) to be modified, preventing other
	// transactions from making changes until this one is complete.
	lockQuery := `
        SELECT device_id FROM public.device_location
        WHERE device_id = $1 AND is_current = TRUE
        FOR UPDATE
    `
	var existingDeviceID string
	err = tx.QueryRowContext(ctx, lockQuery, deviceID).Scan(&existingDeviceID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to lock device_location row: %w", err)
	}

	// Now that we have the lock, proceed with the atomic operations.
	// 1. Mark the old location as not current.
	updateQuery := `
        UPDATE public.device_location
        SET is_current = FALSE
        WHERE device_id = $1 AND is_current = TRUE
    `
	if _, err := tx.ExecContext(ctx, updateQuery, deviceID); err != nil {
		return fmt.Errorf("failed to update old location: %w", err)
	}

	// 2. Insert the new location.
	insertQuery := `
        INSERT INTO public.device_location (device_id, location_id, assigned_at, is_current)
        VALUES ($1, $2, NOW(), TRUE)
    `
	if _, err := tx.ExecContext(ctx, insertQuery, deviceID, locationID); err != nil {
		return fmt.Errorf("failed to insert new location: %w", err)
	}

	// Commit the transaction. If this fails, the deferred rollback is called.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLocationByDeviceID retrieves the location associated with a device ID.
func (d *PostgresDeviceDAO) GetLocationByDeviceID(id string) (*models.Location, error) {
	stmt := `
		SELECT l.location_id, l.location_name
		FROM device_location dl
		JOIN locations l ON dl.location_id = l.location_id
		WHERE dl.device_id = $1 AND dl.is_current = true
	`
	row := d.db.QueryRow(stmt, id)
	var location models.Location
	err := row.Scan(&location.ID, &location.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error scanning location for device %s: %w", id, err)
	}
	return &location, nil
}

// CountAll counts total devices based on search criteria.
func (d *PostgresDeviceDAO) CountAll(ctx context.Context, search string) (int, error) {
	var count int
	var query strings.Builder
	args := []interface{}{}
	argCount := 1

	query.WriteString("SELECT COUNT(d.device_id) FROM devices d ")

	if search != "" {
		query.WriteString(`
			WHERE
				LOWER(d.device_id) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%') OR
				EXISTS (
					SELECT 1
					FROM components c
					WHERE c.device_id = d.device_id AND LOWER(c.component_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
				) OR
				EXISTS (
					SELECT 1
					FROM device_location dl
					JOIN locations l ON dl.location_id = l.location_id
					WHERE dl.device_id = d.device_id AND dl.is_current = true AND LOWER(l.location_name) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
				)
		`)
		args = append(args, search)
		argCount++
	}

	err := d.db.QueryRowContext(ctx, query.String(), args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count devices: %w", err)
	}
	return count, nil
}

// CheckDeviceAccess checks if a user has access to a specific device.
func (d *PostgresDeviceDAO) CheckDeviceAccess(userID int, deviceID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM devices AS d
			JOIN device_location AS dl ON d.device_id = dl.device_id AND dl.is_current = TRUE
			JOIN locations AS l ON dl.location_id = l.location_id
			JOIN user_site AS us ON l.site_id = us.site_id
			WHERE us.user_id = $1 AND d.device_id = $2
		)
	`
	var hasAccess bool
	err := d.db.QueryRow(query, userID, deviceID).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check user access to device: %w", err)
	}
	return hasAccess, nil
}

// UpdateDeviceConsumption updates the current, voltage, and power consumption of a device within a transaction.
func (d *PostgresDeviceDAO) UpdateDeviceConsumption(tx *sql.Tx, id string, current *float64, voltage *float64, power *float64) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE devices SET current = $1, voltage = $2, power = $3 WHERE device_id = $4", current, voltage, power, id)
	if err != nil {
		return fmt.Errorf("failed to update device consumption: %w", err)
	}
	return nil
}

// UpdateDeviceProvisioningToken updates the provisioning token for a device.
func (d *PostgresDeviceDAO) UpdateDeviceProvisioningToken(id string, token string) error {
	_, err := d.db.Exec("UPDATE devices SET provisioning_token = $1 WHERE device_id = $2", token, id)
	if err != nil {
		return fmt.Errorf("failed to update device provisioning token: %w", err)
	}
	return nil
}

// CheckDeviceProvisioningToken checks if the provided provisioning token matches the stored token for a device.
func (d *PostgresDeviceDAO) CheckDeviceProvisioningToken(deviceID string, token string) (bool, error) {
	var storedToken sql.NullString
	err := d.db.QueryRow("SELECT provisioning_token FROM devices WHERE device_id = $1", deviceID).Scan(&storedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Device not found
		}
		return false, fmt.Errorf("failed to retrieve provisioning token: %w", err)
	}

	if !storedToken.Valid || storedToken.String != token {
		return false, nil // Token does not match
	}

	return true, nil // Token matches
}

// CheckDeviceToken checks if the provided token matches the stored token for a device.
func (d *PostgresDeviceDAO) CheckDeviceToken(token string, deviceID string) (bool, error) {
	var storedToken sql.NullString
	err := d.db.QueryRow("SELECT provisioning_token FROM devices WHERE device_id = $1", deviceID).Scan(&storedToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Device not found
		}
		return false, fmt.Errorf("failed to retrieve device token: %w", err)
	}

	if !storedToken.Valid || storedToken.String != token {
		return false, nil // Token does not match
	}

	return true, nil // Token matches
}

// CheckDeviceLocation checks if a device is assigned to a specific location.
func (d *PostgresDeviceDAO) CheckDeviceLocation(deviceID string, locationID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM device_location
			WHERE device_id = $1 AND location_id = $2 AND is_current = true
		)
	`
	err := d.db.QueryRow(query, deviceID, locationID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking device location: %w", err)
	}
	return exists, nil
}
