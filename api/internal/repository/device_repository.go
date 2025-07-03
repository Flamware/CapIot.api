package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DeviceDAO defines the interface for all device and related data access operations.
type DeviceDAO interface {
	CreateDevice(device *models.Device) error
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(deviceID string) error
	CreateCaptor(captor *models.Captor) (*models.Captor, error)
	GetCaptorByID(id string) (*models.Captor, error)
	GetAllDevices() ([]*models.Device, error) // Ensure this
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	InsertDeviceCaptor(captor *models.DeviceCaptor) error
	UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetCaptorsByDeviceID(id string) ([]*models.Captor, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
	UnassignDeviceFromLocation(id string) error
	FindAllWithSensorsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithSensorsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
	UpdateCaptor(captor *models.Captor) error
	UpdateCaptorRange(captor *models.Captor) error
}

// PostgresDeviceDAO implements the DeviceDAO interface using PostgreSQL.
type PostgresDeviceDAO struct {
	db *sql.DB
}

// NewPostgresDeviceDAO creates a new PostgresDeviceDAO instance.
func NewPostgresDeviceDAO(db *sql.DB) *PostgresDeviceDAO {
	return &PostgresDeviceDAO{db: db}
}

// Implementations for Device operations
func (d *PostgresDeviceDAO) InsertDevice(device *models.Device) error {
	_, err := d.db.Exec("INSERT INTO devices (device_id, last_seen, status, created_at) VALUES ($1, NOW(), $2, NOW())", device.DeviceID, device.Status)
	return err
}

// GetAllDevices retrieves all device records from the database
func (d *PostgresDeviceDAO) GetAllDevices() ([]*models.Device, error) { // Changed return type to []*models.Device
	rows, err := d.db.Query("SELECT device_id, last_seen, status, created_at FROM devices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device // Changed to slice of pointers
	for rows.Next() {
		var dev models.Device
		if err := rows.Scan(&dev.DeviceID, &dev.LastSeen, &dev.Status, &dev.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, &dev) // Append a pointer to the device
	}
	return devices, nil
}

func (d *PostgresDeviceDAO) GetDeviceByDeviceID(id string) (*models.Device, error) {
	row := d.db.QueryRow("SELECT device_id, last_seen, status, created_at FROM devices WHERE device_id = $1", id)
	var device models.Device
	err := row.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (d *PostgresDeviceDAO) UpdateDeviceLastSeenAndStatus(deviceID string) error {
	_, err := d.db.Exec("UPDATE devices SET last_seen = NOW(), status = $1 WHERE device_id = $2", "stopped", deviceID) // Assuming status becomes online on availability
	return err
}

func (d *PostgresDeviceDAO) DeviceExists(deviceID string) (bool, error) {
	var exists bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM devices WHERE device_id = $1)", deviceID).Scan(&exists)
	return exists, err
}

func (d *PostgresDeviceDAO) IsDeviceAssigned(deviceID string) (bool, error) {
	var assigned bool
	err := d.db.QueryRow("SELECT EXISTS(SELECT 1 FROM device_location WHERE device_id = $1 AND is_current = true)", deviceID).Scan(&assigned)
	return assigned, err
}

func (d *PostgresDeviceDAO) SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO device_location (device_id, location_id, assigned_at, is_current) VALUES ($1, $2, NOW(), true)", deviceID, locationID)
	if err != nil {
		// Log the error with context. This is crucial for debugging.
		fmt.Printf("Error setting device '%s' to location '%d': %v\n", deviceID, locationID, err)
		// Return the error. The calling service layer will decide what to do with it.
		return fmt.Errorf("failed to set device '%s' to location '%d': %w", deviceID, locationID, err)
	}
	return nil // Return nil to indicate success
}
func (d *PostgresDeviceDAO) InsertCaptor(captor *models.Captor) (*models.Captor, error) {
	err := d.db.QueryRow("INSERT INTO captors (captor_type, captor_type) VALUES ($1, $2) RETURNING captor_id", captor.CaptorType).Scan(&captor.CaptorID)
	if err != nil {
		return nil, err
	}
	return captor, nil
}

// Implementations for DeviceCaptor operations
func (d *PostgresDeviceDAO) InsertDeviceCaptor(dc *models.DeviceCaptor) error {
	_, err := d.db.Exec("INSERT INTO public.device_captors (device_id, captor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", dc.DeviceID, dc.CaptorID)
	return err
}

// GetCaptorByID retrieves a captor record by its ID (string)
func (d *PostgresDeviceDAO) GetCaptorByID(id string) (*models.Captor, error) {
	stmt := `
       SELECT captor_id, captor_type
       FROM captors
       WHERE captor_id = $1
    `
	row := d.db.QueryRow(stmt, id)
	var captor models.Captor
	err := row.Scan(&captor.CaptorID, &captor.CaptorType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &captor, nil
}

func (d *PostgresDeviceDAO) CreateCaptor(captor *models.Captor) (*models.Captor, error) {
	stmt := `
		INSERT INTO captors (captor_id, captor_type)
		VALUES ($1, $2)
		RETURNING captor_id
	`
	err := d.db.QueryRow(stmt, captor.CaptorID, captor.CaptorType).Scan(&captor.CaptorID)
	if err != nil {
		return nil, err
	}
	return captor, nil
}
func (d *PostgresDeviceDAO) CreateDevice(device *models.Device) error {
	_, err := d.db.Exec(
		"INSERT INTO devices (device_id, last_seen, status, created_at) VALUES ($1, $2, $3, $4)",
		device.DeviceID,
		device.LastSeen,
		device.Status,
		device.CreatedAt,
	)
	return err
}

func (d *PostgresDeviceDAO) UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error {
	_, err := d.db.Exec("UPDATE devices SET status = $1 WHERE device_id = $2", status, id)
	return err
}

func (d *PostgresDeviceDAO) GetDeviceByID(id string) (*models.Device, error) {
	row := d.db.QueryRow("SELECT device_id, last_seen, status, created_at FROM devices WHERE device_id = $1", id)
	var device models.Device
	err := row.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return &models.Device{}, nil // Return a pointer to an empty Device struct
		}
		return nil, err
	}
	return &device, nil
}

func (d *PostgresDeviceDAO) GetCaptorsByDeviceID(id string) ([]*models.Captor, error) {
	rows, err := d.db.Query(`
        SELECT c.captor_id, c.captor_type
        FROM device_captors dc
        JOIN captors c ON dc.captor_id = c.captor_id
        WHERE dc.device_id = $1
    `, id)
	if err != nil {
		return []*models.Captor{}, err
	}
	defer rows.Close()

	var captors []*models.Captor
	for rows.Next() {
		var captor models.Captor
		if err := rows.Scan(&captor.CaptorID, &captor.CaptorType); err != nil {
			return []*models.Captor{}, err
		}
		captors = append(captors, &captor)
	}

	if err := rows.Err(); err != nil {
		return []*models.Captor{}, err
	}

	return captors, nil
}

func (d *PostgresDeviceDAO) GetUnassignedDevices() ([]*models.Device, error) {
	rows, err := d.db.Query(`
       SELECT d.device_id, d.last_seen, d.status, d.created_at
       FROM devices d
       LEFT JOIN device_location dl ON d.device_id = dl.device_id
       WHERE dl.device_id IS NULL
    `)
	if err != nil {
		return []*models.Device{}, err // Return empty slice on query error
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var device models.Device
		if err := rows.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt); err != nil {
			return []*models.Device{}, err // Return empty slice on scan error
		}
		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return []*models.Device{}, err // Return empty slice on rows.Err()
	}

	return devices, nil // Return the (potentially empty) slice of devices
}

func (d *PostgresDeviceDAO) DeleteDevice(id string) error {
	_, err := d.db.Exec("DELETE FROM devices WHERE device_id = $1", id)
	return err
}

func (d *PostgresDeviceDAO) UnassignDeviceFromLocation(id string) error {
	_, err := d.db.Exec("UPDATE device_location SET is_current = false WHERE device_id = $1", id)
	return err
}

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
			return nil, nil // No location found for this device
		}
		return nil, err
	}
	return &location, nil
}
func (d *PostgresDeviceDAO) FindAllWithSensorsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithSensorsAndLocation, error) {
	var devicesWithInfo []*models.DeviceWithSensorsAndLocation
	var query strings.Builder
	args := []interface{}{limit, offset}
	argCount := 3

	query.WriteString(`
       SELECT
          d.device_id, d.last_seen, d.status, d.created_at,
          l.location_id, l.location_name
       FROM devices d
       LEFT JOIN device_location dl ON d.device_id = dl.device_id AND dl.is_current = true
       LEFT JOIN locations l ON dl.location_id = l.location_id
    `)

	if search != "" {
		query.WriteString(`
          WHERE
             LOWER(d.device_id) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%') OR
             EXISTS (
                SELECT 1
                FROM device_captors dc
                JOIN captors c ON dc.captor_id = c.captor_id
                WHERE dc.device_id = d.device_id AND LOWER(c.captor_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
             ) OR
             LOWER(l.location_name) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
       `)
		args = append(args, search)
		argCount++
	}

	query.WriteString(`
       LIMIT $1
       OFFSET $2
    `)

	rows, err := d.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	deviceMap := make(map[string]*models.DeviceWithSensorsAndLocation)

	for rows.Next() {
		var deviceWithInfo models.DeviceWithSensorsAndLocation
		var location models.Location
		var device models.Device // Still need to scan into a Device struct
		if err := rows.Scan(
			&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt,
			&location.ID, &location.Name,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		deviceWithCaptors := &models.DeviceWithCaptors{
			Device: &device,
		}

		if existingDevice, ok := deviceMap[device.DeviceID]; ok {
			existingDevice.Location = &location
		} else {
			deviceWithInfo.DeviceWithCaptors = deviceWithCaptors // Assign DeviceWithCaptors
			deviceWithInfo.Location = &location
			ptr := &deviceWithInfo
			deviceMap[device.DeviceID] = ptr
			devicesWithInfo = append(devicesWithInfo, ptr)
		}
	}

	if err := rows.Err(); err != nil { // Check for errors during row iteration
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// Fetch captors for each device
	for _, deviceInfo := range devicesWithInfo {
		if deviceInfo.DeviceWithCaptors != nil && deviceInfo.DeviceWithCaptors.Device != nil {
			captors, err := d.GetCaptorsByDeviceID(deviceInfo.DeviceWithCaptors.Device.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get captors for device %s: %w", deviceInfo.DeviceWithCaptors.Device.DeviceID, err)
			}
			deviceInfo.DeviceWithCaptors.Captors = captors // Assign captors to DeviceWithCaptors
		}
	}

	return devicesWithInfo, nil // Will return an empty slice if no devices were found
}

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
					FROM device_captors dc
					JOIN captors c ON dc.captor_id = c.captor_id
					WHERE dc.device_id = d.device_id AND LOWER(c.captor_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
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

func (d *PostgresDeviceDAO) UpdateCaptor(captor *models.Captor) error {
	_, err := d.db.Exec("UPDATE captors SET captor_type = $1 WHERE captor_id = $2", captor.CaptorType, captor.CaptorID)
	if err != nil {
		return fmt.Errorf("failed to update captor: %w", err)
	}
	return nil
}

func (d *PostgresDeviceDAO) UpdateCaptorRange(captor *models.Captor) error {
	_, err := d.db.Exec("UPDATE captors SET min_threshold = $1, max_threshold = $2 WHERE captor_id = $3", captor.MinThreshold, captor.MaxThreshold, captor.CaptorID)
	if err != nil {
		return fmt.Errorf("failed to update captor range: %w", err)
	}
	return nil
}
