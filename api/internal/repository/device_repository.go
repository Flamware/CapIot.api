package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time" // Added for time.Now()
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
		ON CONFLICT (device_id) DO NOTHING` // Using DO NOTHING for idempotency
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
	row := d.db.QueryRow("SELECT device_id, last_seen, status, created_at FROM devices WHERE device_id = $1", id)
	var device models.Device
	err := row.Scan(&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt)
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
	rows, err := d.db.Query("SELECT device_id, last_seen, status, created_at FROM devices")
	if err != nil {
		return nil, fmt.Errorf("failed to query all devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var dev models.Device
		if err := rows.Scan(&dev.DeviceID, &dev.LastSeen, &dev.Status, &dev.CreatedAt); err != nil {
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
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok || tx == nil {
		return fmt.Errorf("transaction not found in context")
	}
	executor := d.getExecutor(tx)

	// First, mark any existing assignments for this device as not current
	updateQuery := `
		UPDATE public.device_location
		SET is_current = FALSE
		WHERE device_id = $1 AND is_current = TRUE
	`
	_, err := executor.Exec(updateQuery, deviceID)
	if err != nil {
		return fmt.Errorf("failed to mark old device location as not current: %w", err)
	}

	// Then, insert the new assignment or update an existing one to be current
	insertQuery := `
		INSERT INTO public.device_location (device_id, location_id, assigned_at, is_current)
		VALUES ($1, $2, $3, TRUE)
		ON CONFLICT (device_id, location_id) DO UPDATE
		SET assigned_at = $3, is_current = TRUE
	`
	_, err = executor.Exec(insertQuery, deviceID, locationID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set device %s to location %d: %w", deviceID, locationID, err)
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

// FindAllWithComponentsAndLocations retrieves devices with their components, locations, and site names.
func (d *PostgresDeviceDAO) FindAllWithComponentsAndLocations(
	ctx context.Context,
	limit int,
	offset int,
	search string,
) ([]*models.DeviceWithComponentsAndLocation, error) {
	query := `
		SELECT
			d.device_id, d.last_seen, d.status, d.created_at,
			l.location_id, l.location_name, l.location_description, l.site_id,
			s.site_name
		FROM devices d
		LEFT JOIN device_location dl ON d.device_id = dl.device_id AND dl.is_current = true
		LEFT JOIN locations l ON dl.location_id = l.location_id
		LEFT JOIN sites s ON l.site_id = s.site_id
	`
	args := []interface{}{}
	argIndex := 1

	// Add search filter
	if search != "" {
		query += fmt.Sprintf(`
			WHERE LOWER(d.device_id) LIKE LOWER('%%' || $%d || '%%')
			OR LOWER(l.location_name) LIKE LOWER('%%' || $%d || '%%')
			OR LOWER(s.site_name) LIKE LOWER('%%' || $%d || '%%')
		`, argIndex, argIndex, argIndex)
		args = append(args, search)
		argIndex++
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY d.device_id LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query for devices with locations and site names failed: %w", err)
	}
	defer rows.Close()

	var devices []*models.DeviceWithComponentsAndLocation

	for rows.Next() {
		var device models.Device
		var location models.LocationWithSite

		var locationID, siteID sql.NullInt64
		var locationName, locationDescription, siteName sql.NullString

		if err := rows.Scan(
			&device.DeviceID, &device.LastSeen, &device.Status, &device.CreatedAt,
			&locationID, &locationName, &locationDescription, &siteID,
			&siteName,
		); err != nil {
			return nil, fmt.Errorf("scan device row failed: %w", err)
		}

		deviceInfo := &models.DeviceWithComponentsAndLocation{
			DeviceWithComponents: &models.DeviceWithComponents{
				Device: &device,
			},
		}

		// Assign location
		if locationID.Valid {
			id := int(locationID.Int64)
			location.ID = &id
			if locationName.Valid {
				location.Name = &locationName.String
			}
			if locationDescription.Valid {
				location.Description = &locationDescription.String
			}
			if siteID.Valid {
				sID := int(siteID.Int64)
				location.SiteID = &sID
			}
			if siteName.Valid {
				location.SiteName = &siteName.String
			}
			deviceInfo.Location = &location
		}

		devices = append(devices, deviceInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Fetch components for each device
	for _, deviceInfo := range devices {
		if deviceInfo.DeviceWithComponents != nil && deviceInfo.DeviceWithComponents.Device != nil {
			components, err := d.GetcomponentsByDeviceID(deviceInfo.DeviceWithComponents.Device.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get components for device %s: %w", deviceInfo.DeviceWithComponents.Device.DeviceID, err)
			}
			deviceInfo.DeviceWithComponents.Components = components
		}
	}

	return devices, nil
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
                FROM device_components dc
                JOIN components c ON dc.component_id = c.component_id
                WHERE dc.device_id = d.device_id AND LOWER(c.component_type) LIKE LOWER('%' || $` + fmt.Sprintf("%d", argCount) + ` || '%')
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

// --- Component Operations ---

// CreateComponent creates a new component instance within a transaction.
func (d *PostgresDeviceDAO) CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error) {
	executor := d.getExecutor(tx)
	var minThreshold, maxThreshold sql.NullFloat64
	if component.MinThreshold != nil {
		minThreshold = sql.NullFloat64{Float64: *component.MinThreshold, Valid: true}
	}
	var maxRunningHours sql.NullInt32
	if component.MaxRunningHours != nil {
		maxRunningHours = sql.NullInt32{Int32: *component.MaxRunningHours, Valid: true}
	}

	query := `
        INSERT INTO public.components (
            component_id, component_name, component_type, component_subtype,
            component_status, min_threshold, max_threshold, max_running_hours
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING component_id, component_name, component_type, component_subtype,
                  component_status, min_threshold, max_threshold, max_running_hours
    `
	err := executor.QueryRow(
		query,
		component.ComponentID,
		component.ComponentName,
		component.ComponentType,
		component.ComponentSubtype,
		component.ComponentStatus,
		minThreshold,
		maxThreshold,
		maxRunningHours,
	).Scan(
		&component.ComponentID,
		&component.ComponentName,
		&component.ComponentType,
		&component.ComponentSubtype,
		&component.ComponentStatus,
		&minThreshold,
		&maxThreshold,
		&maxRunningHours,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	if minThreshold.Valid {
		component.MinThreshold = &minThreshold.Float64
	} else {
		component.MinThreshold = nil
	}
	if maxThreshold.Valid {
		component.MaxThreshold = &maxThreshold.Float64
	} else {
		component.MaxThreshold = nil
	}
	if maxRunningHours.Valid {
		component.MaxRunningHours = &maxRunningHours.Int32
	} else {
		component.MaxRunningHours = nil
	}

	return component, nil
}

// GetComponentByID retrieves a component record by its ID.
func (d *PostgresDeviceDAO) GetComponentByID(id string) (*models.Component, error) {
	stmt := `
       SELECT component_id, component_name, component_type, component_subtype,
              component_status, min_threshold, max_threshold, max_running_hours
       FROM components
       WHERE component_id = $1
    `
	row := d.db.QueryRow(stmt, id)
	var component models.Component
	var minThreshold, maxThreshold sql.NullFloat64
	var maxRunningHours sql.NullInt32

	err := row.Scan(
		&component.ComponentID,
		&component.ComponentName,
		&component.ComponentType,
		&component.ComponentSubtype,
		&component.ComponentStatus,
		&minThreshold,
		&maxThreshold,
		&maxRunningHours,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("error scanning component with ID %s: %w", id, err)
	}

	if minThreshold.Valid {
		component.MinThreshold = &minThreshold.Float64
	}
	if maxThreshold.Valid {
		component.MaxThreshold = &maxThreshold.Float64
	}
	if maxRunningHours.Valid {
		component.MaxRunningHours = &maxRunningHours.Int32
	}

	return &component, nil
}

// LinkComponentToDevice links a component to a device within a transaction.
func (d *PostgresDeviceDAO) LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error {
	executor := d.getExecutor(tx)
	query := `
		INSERT INTO public.device_components (device_id, component_id, installation_date)
		VALUES ($1, $2, $3)
		ON CONFLICT (device_id, component_id) DO UPDATE SET removal_date = NULL, installation_date = $3
	` // ON CONFLICT ensures idempotency and handles re-installation
	_, err := executor.Exec(query, deviceID, componentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to link component %s to device %s: %w", componentID, deviceID, err)
	}
	return nil
}

// UpdateComponentStatus updates the status of a component within a transaction.
func (d *PostgresDeviceDAO) UpdateComponentStatus(tx *sql.Tx, id string, status string) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE components SET component_status = $1 WHERE component_id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update component status for %s: %w", id, err)
	}
	return nil
}

// GetcomponentsByDeviceID retrieves all components linked to a specific device.
func (d *PostgresDeviceDAO) GetcomponentsByDeviceID(id string) ([]*models.Component, error) {
	rows, err := d.db.Query(`
        SELECT c.component_id, c.component_name, c.component_type, c.component_subtype,
               c.component_status, c.min_threshold, c.max_threshold, c.max_running_hours,
               c.current_running_hours
        FROM public.components c
        JOIN public.device_components dc ON c.component_id = dc.component_id
        WHERE dc.device_id = $1 AND dc.removal_date IS NULL
    `, id)
	if err != nil {
		return nil, fmt.Errorf("error querying components for device %s: %w", id, err)
	}
	defer rows.Close()

	var components []*models.Component
	for rows.Next() {
		var component models.Component
		var minThreshold, maxThreshold sql.NullFloat64
		var maxRunningHours sql.NullInt32

		err := rows.Scan(
			&component.ComponentID,
			&component.ComponentName,
			&component.ComponentType,
			&component.ComponentSubtype,
			&component.ComponentStatus,
			&minThreshold,
			&maxThreshold,
			&maxRunningHours,
			&component.CurrentRunningHours,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning component row for device %s: %w", id, err)
		}

		if minThreshold.Valid {
			component.MinThreshold = &minThreshold.Float64
		}
		if maxThreshold.Valid {
			component.MaxThreshold = &maxThreshold.Float64
		}
		if maxRunningHours.Valid {
			component.MaxRunningHours = &maxRunningHours.Int32
		}
		components = append(components, &component)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating component rows for device %s: %w", id, err)
	}

	return components, nil
}

// UpdateComponentRange updates the operational range of a component within a transaction.
func (d *PostgresDeviceDAO) UpdateComponentRange(tx *sql.Tx, component *models.Component) error {
	executor := d.getExecutor(tx)
	var minThreshold, maxThreshold sql.NullFloat64
	if component.MinThreshold != nil {
		minThreshold = sql.NullFloat64{Float64: *component.MinThreshold, Valid: true}
	}
	if component.MaxThreshold != nil {
		maxThreshold = sql.NullFloat64{Float64: *component.MaxThreshold, Valid: true}
	}

	_, err := executor.Exec("UPDATE components SET min_threshold = $1, max_threshold = $2 WHERE component_id = $3", minThreshold, maxThreshold, component.ComponentID)
	if err != nil {
		return fmt.Errorf("failed to update component range for %s: %w", component.ComponentID, err)
	}
	return nil
}

// --- Log Operations ---

// HandleDeviceAlert inserts a new alert log for a component within a transaction.
func (d *PostgresDeviceDAO) HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error {
	executor := d.getExecutor(tx)
	stmt := `
        INSERT INTO public.component_log (component_id, log_content, log_read)
        VALUES ($1, $2, FALSE)
    `
	_, err := executor.Exec(stmt, componentID, message)
	if err != nil {
		return fmt.Errorf("failed to insert alert log for component '%s': %w", componentID, err)
	}
	return nil
}

// GetcomponentLogsByComponentID retrieves all logs for a specific component.
func (d *PostgresDeviceDAO) GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error) {
	rows, err := d.db.Query("SELECT log_id, component_id, log_timestamp, log_content, log_read FROM component_log WHERE component_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to query component logs by ID %s: %w", id, err)
	}
	defer rows.Close()

	var logs []*models.ComponentLog
	for rows.Next() {
		var log models.ComponentLog
		if err := rows.Scan(&log.LogID, &log.ComponentID, &log.Timestamp, &log.Content, &log.Read); err != nil {
			return nil, fmt.Errorf("failed to scan component log row: %w", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during component log rows iteration: %w", err)
	}
	return logs, nil
}

// GetcomponentLogsByDeviceIDAndComponentID retrieves component logs for a specific device and component.
func (d *PostgresDeviceDAO) GetcomponentLogsByDeviceIDAndComponentID(deviceID string, ComponentID string) ([]*models.ComponentLog, error) {
	rows, err := d.db.Query(`
       SELECT sl.log_id, sl.component_id, sl.log_timestamp, sl.log_content, sl.log_read
       FROM component_log sl
       JOIN device_components ds ON sl.component_id = ds.component_id
       WHERE ds.device_id = $1 AND ds.component_id = $2
    `, deviceID, ComponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query component logs by device and component ID: %w", err)
	}
	defer rows.Close()

	var logs []*models.ComponentLog
	for rows.Next() {
		var log models.ComponentLog
		if err := rows.Scan(&log.LogID, &log.ComponentID, &log.Timestamp, &log.Content, &log.Read); err != nil {
			return nil, fmt.Errorf("failed to scan component log row by device and component ID: %w", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during component log rows iteration by device and component ID: %w", err)
	}
	return logs, nil
}

// GetDeviceLogsByDeviceID retrieves all logs for a specific device.
func (d *PostgresDeviceDAO) GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error) {
	rows, err := d.db.Query(`
       SELECT sl.log_id, sl.component_id, sl.log_timestamp, sl.log_content, sl.log_read
       FROM component_log sl
       JOIN device_components ds ON sl.component_id = ds.component_id
       WHERE ds.device_id = $1
    `, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query device logs by device ID: %w", err)
	}
	defer rows.Close()

	var logs []*models.ComponentLog
	for rows.Next() {
		var log models.ComponentLog
		if err := rows.Scan(&log.LogID, &log.ComponentID, &log.Timestamp, &log.Content, &log.Read); err != nil {
			return nil, fmt.Errorf("failed to scan device log row by device ID: %w", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during device log rows iteration by device ID: %w", err)
	}
	return logs, nil
}

// GetAllLogsByUser retrieves all logs for the user.
func (d *PostgresDeviceDAO) GetAllLogsByUser(userID int) ([]*models.ComponentLog, error) {
	rows, err := d.db.Query(`
        SELECT sl.log_id, sl.component_id, sl.log_timestamp, sl.log_content, sl.log_read
        FROM component_log sl
        JOIN device_components ds ON sl.component_id = ds.component_id
        JOIN device_location dl ON ds.device_id = dl.device_id AND dl.is_current = true
        JOIN locations l ON dl.location_id = l.location_id
        JOIN user_site us ON l.site_id = us.site_id
        WHERE us.user_id = $1
    `, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to query all logs by user: %w", err)
	}
	defer rows.Close()

	var logs []*models.ComponentLog
	for rows.Next() {
		var log models.ComponentLog
		if err := rows.Scan(&log.LogID, &log.ComponentID, &log.Timestamp, &log.Content, &log.Read); err != nil {
			return nil, fmt.Errorf("failed to scan component log row for all logs by user: %w", err)
		}
		logs = append(logs, &log)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during all logs by user rows iteration: %w", err)
	}
	return logs, nil
}
func (d *PostgresDeviceDAO) UserHasAccessTocomponent(userID int, ComponentID string) (bool, error) {
	query := `
       SELECT EXISTS (
          SELECT 1
          FROM device_components AS dc
          JOIN devices AS d ON dc.device_id = d.device_id
          JOIN device_location AS dl ON d.device_id = dl.device_id AND dl.is_current = TRUE
          JOIN locations AS l ON dl.location_id = l.location_id
          JOIN user_site AS us ON l.site_id = us.site_id
          WHERE us.user_id = $1 AND dc.component_id = $2
       )
    `
	var hasAccess bool
	err := d.db.QueryRow(query, userID, ComponentID).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check user access to component: %w", err)
	}
	return hasAccess, nil
}

// MarkcomponentLogsAsRead marks component logs as read for a specific component within a transaction.
func (d *PostgresDeviceDAO) MarkcomponentLogsAsRead(tx *sql.Tx, ComponentID string, logIds []int) error {
	if len(logIds) == 0 {
		return nil // No logs to mark as read
	}

	executor := d.getExecutor(tx)
	placeholders := make([]string, len(logIds))
	args := make([]interface{}, len(logIds)+1)
	for i, id := range logIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(logIds)] = ComponentID // Last argument is the component ID

	query := fmt.Sprintf(`
       UPDATE component_log
       SET log_read = TRUE
       WHERE log_id IN (%s) AND component_id = $%d
    `, strings.Join(placeholders, ", "), len(args))

	_, err := executor.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark component logs as read: %w", err)
	}
	return nil
}

// MarkAllLogsAsRead marks all logs as read for a specific user within a transaction.
func (d *PostgresDeviceDAO) MarkAllLogsAsRead(tx *sql.Tx, userID int) error {
	executor := d.getExecutor(tx)
	query := `
       UPDATE component_log AS cl
       SET log_read = TRUE
       FROM device_components AS dc
       JOIN device_location AS dl ON dc.device_id = dl.device_id AND dl.is_current = TRUE
       JOIN locations AS l ON dl.location_id = l.location_id
       JOIN user_site AS us ON l.site_id = us.site_id
       WHERE us.user_id = $1 AND cl.component_id = dc.component_id
    `
	_, err := executor.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all logs as read for user %d: %w", userID, err)
	}
	return nil
}

// UpdateComponentRunningHours updates the running hours of a component within a transaction.
func (d *PostgresDeviceDAO) UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error {
	executor := d.getExecutor(tx)
	_, err := executor.Exec("UPDATE components SET current_running_hours = $1 WHERE component_id = $2", hours, id)
	if err != nil {
		return fmt.Errorf("failed to update component running hours for %s: %w", id, err)
	}
	return nil
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

// GetSensorsByDeviceID retrieves all sensor components linked to a specific device.
func (d *PostgresDeviceDAO) GetSensorsByDeviceID(id string) ([]*models.Component, error) {
	rows, err := d.db.Query(`
		SELECT c.component_id, c.component_name, c.component_type, c.component_subtype,
		       			   c.component_status, c.min_threshold, c.max_threshold, c.max_running_hours,
		       			   c.current_running_hours
		FROM public.components c
		JOIN public.device_components dc ON c.component_id = dc.component_id
		WHERE dc.device_id = $1 AND dc.removal_date IS NULL AND c.component_type = 'sensor'
	`, id)
	if err != nil {
		return nil, fmt.Errorf("error querying sensor components for device %s: %w", id, err)
	}
	defer rows.Close()

	var components []*models.Component
	for rows.Next() {
		var component models.Component
		var minThreshold, maxThreshold sql.NullFloat64
		var maxRunningHours sql.NullInt32

		err := rows.Scan(
			&component.ComponentID,
			&component.ComponentName,
			&component.ComponentType,
			&component.ComponentSubtype,
			&component.ComponentStatus,
			&minThreshold,
			&maxThreshold,
			&maxRunningHours,
			&component.CurrentRunningHours,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning sensor component row for device %s: %w", id, err)
		}
		if minThreshold.Valid {
			component.MinThreshold = &minThreshold.Float64
		}
		if maxThreshold.Valid {
			component.MaxThreshold = &maxThreshold.Float64
		}
		if maxRunningHours.Valid {
			component.MaxRunningHours = &maxRunningHours.Int32
		}
		components = append(components, &component)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sensor component rows for device %s: %w", id, err)
	}

	return components, nil
}
