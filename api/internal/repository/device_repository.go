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
	Createsensor(sensor *models.Sensor) (*models.Sensor, error)
	GetsensorByID(id string) (*models.Sensor, error)
	GetAllDevices() ([]*models.Device, error) // Ensure this
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	InsertDevicesensor(sensor *models.Devicesensor) error
	UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetsensorsByDeviceID(id string) ([]*models.Sensor, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
	UnassignDeviceFromLocation(id string) error
	FindAllWithSensorsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithSensorsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
	Updatesensor(sensor *models.Sensor) error
	UpdatesensorRange(sensor *models.Sensor) error
	HandleDeviceAlert(SensorID string, message string) error
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
func (d *PostgresDeviceDAO) Insertsensor(sensor *models.Sensor) (*models.Sensor, error) {
	err := d.db.QueryRow("INSERT INTO sensors (sensor_type, sensor_type) VALUES ($1, $2) RETURNING sensor_id", sensor.SensorType).Scan(&sensor.SensorID)
	if err != nil {
		return nil, err
	}
	return sensor, nil
}

// Implementations for Devicesensor operations
func (d *PostgresDeviceDAO) InsertDevicesensor(dc *models.Devicesensor) error {
	_, err := d.db.Exec("INSERT INTO public.device_sensors (device_id, sensor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", dc.DeviceID, dc.SensorID)
	return err
}

// GetsensorByID retrieves a sensor record by its ID (string)
func (d *PostgresDeviceDAO) GetsensorByID(id string) (*models.Sensor, error) {
	stmt := `
       SELECT sensor_id, sensor_type, min_threshold, max_threshold  -- <--- Selecting 4 columns
       FROM sensors
       WHERE sensor_id = $1
    `
	row := d.db.QueryRow(stmt, id)
	var sensor models.Sensor
	err := row.Scan(&sensor.SensorID, &sensor.SensorType, &sensor.MinThreshold, &sensor.MaxThreshold)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sensor, nil
}

func (d *PostgresDeviceDAO) Createsensor(sensor *models.Sensor) (*models.Sensor, error) {
	stmt := `
		INSERT INTO sensors (sensor_id, sensor_type)
		VALUES ($1, $2)
		RETURNING sensor_id
	`
	err := d.db.QueryRow(stmt, sensor.SensorID, sensor.SensorType).Scan(&sensor.SensorID)
	if err != nil {
		return nil, err
	}
	return sensor, nil
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

func (d *PostgresDeviceDAO) GetsensorsByDeviceID(id string) ([]*models.Sensor, error) {
	rows, err := d.db.Query(`
        SELECT c.sensor_id, c.sensor_type, c.min_threshold, c.max_threshold
        FROM device_sensors dc
        JOIN sensors c ON dc.sensor_id = c.sensor_id
        WHERE dc.device_id = $1
    `, id)
	if err != nil {
		return []*models.Sensor{}, err
	}
	defer rows.Close()

	var sensors []*models.Sensor
	for rows.Next() {
		var sensor models.Sensor
		if err := rows.Scan(&sensor.SensorID, &sensor.SensorType, &sensor.MinThreshold, &sensor.MaxThreshold); err != nil {
			return []*models.Sensor{}, err
		}
		sensors = append(sensors, &sensor)
	}

	if err := rows.Err(); err != nil {
		return []*models.Sensor{}, err
	}

	return sensors, nil
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
                FROM device_sensors dc
                JOIN sensors c ON dc.sensor_id = c.sensor_id
                WHERE dc.device_id = d.device_id AND LOWER(c.sensor_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
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

		deviceWithsensors := &models.DeviceWithsensors{
			Device: &device,
		}

		if existingDevice, ok := deviceMap[device.DeviceID]; ok {
			existingDevice.Location = &location
		} else {
			deviceWithInfo.DeviceWithsensors = deviceWithsensors // Assign DeviceWithsensors
			deviceWithInfo.Location = &location
			ptr := &deviceWithInfo
			deviceMap[device.DeviceID] = ptr
			devicesWithInfo = append(devicesWithInfo, ptr)
		}
	}

	if err := rows.Err(); err != nil { // Check for errors during row iteration
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	// Fetch sensors for each device
	for _, deviceInfo := range devicesWithInfo {
		if deviceInfo.DeviceWithsensors != nil && deviceInfo.DeviceWithsensors.Device != nil {
			sensors, err := d.GetsensorsByDeviceID(deviceInfo.DeviceWithsensors.Device.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get sensors for device %s: %w", deviceInfo.DeviceWithsensors.Device.DeviceID, err)
			}
			deviceInfo.DeviceWithsensors.Sensors = sensors // Assign sensors to DeviceWithsensors
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
					FROM device_sensors dc
					JOIN sensors c ON dc.sensor_id = c.sensor_id
					WHERE dc.device_id = d.device_id AND LOWER(c.sensor_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
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

func (d *PostgresDeviceDAO) Updatesensor(sensor *models.Sensor) error {
	_, err := d.db.Exec("UPDATE sensors SET sensor_type = $1 WHERE sensor_id = $2", sensor.SensorType, sensor.SensorID)
	if err != nil {
		return fmt.Errorf("failed to update sensor: %w", err)
	}
	return nil
}

func (d *PostgresDeviceDAO) UpdatesensorRange(sensor *models.Sensor) error {
	_, err := d.db.Exec("UPDATE sensors SET min_threshold = $1, max_threshold = $2 WHERE sensor_id = $3", sensor.MinThreshold, sensor.MaxThreshold, sensor.SensorID)
	if err != nil {
		return fmt.Errorf("failed to update sensor range: %w", err)
	}
	return nil
}

func (d *PostgresDeviceDAO) HandleDeviceAlert(sensor_id string, message string) error {
	stmt := `
        INSERT INTO public.sensor_log (sensor_id, log_content, log_read)
        VALUES ($1, $2, FALSE) -- You can explicitly set log_read to FALSE here
    `
	// Execute the statement, passing the values as separate arguments to Exec
	_, err := d.db.Exec(stmt, sensor_id, message)
	if err != nil {
		// Wrap the error for more context in the logs
		return fmt.Errorf("failed to insert alert log for sensor '%s': %w", sensor_id, err)
	}
	return nil
}

// Function to return the log content for a specific sensor
func (d *PostgresDeviceDAO) GetSensorLog(sensorID string) ([]*models.SensorLog, error) {
	rows, err := d.db.Query("SELECT log_id, sensor_id, log_content, log_read, created_at FROM sensor_log WHERE sensor_id = $1", sensorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensor logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.SensorLog
	for rows.Next() {
		var log models.SensorLog
		if err := rows.Scan(&log.LogID, &log.SensorID, &log.Content, &log.Read, &log.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan sensor log: %w", err)
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return logs, nil
}
