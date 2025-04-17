package repository

import (
	"CapIot-api/internal/models"
	"database/sql"
)

// DeviceDAO defines the interface for all device and related data access operations.
type DeviceDAO interface {
	CreateDevice(device *models.Device) error
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(deviceID string) error
	CreateCaptor(captor *models.Captor) (*models.Captor, error)
	GetCaptorByID(id string) (*models.Captor, error)
	GetAllDevices() ([]*models.Device, error) // Ensure this
	SetDeviceToLocation(deviceID string, locationID int) error
	InsertDeviceCaptor(captor *models.DeviceCaptor) error
	UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetCaptorsByDeviceID(id string) ([]*models.Captor, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(id string) error
	UnassignDeviceFromLocation(id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
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
	_, err := d.db.Exec("UPDATE devices SET last_seen = NOW(), status = $1 WHERE device_id = $2", "online", deviceID) // Assuming status becomes online on availability
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

func (d *PostgresDeviceDAO) SetDeviceToLocation(deviceID string, locationID int) error {
	// You'll need to handle updating previous current location if needed
	_, err := d.db.Exec("INSERT INTO device_location (device_id, location_id, assigned_at, is_current) VALUES ($1, $2, NOW(), true)", deviceID, locationID)
	return err
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
