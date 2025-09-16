package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// PostgresComponentDAO handles database operations for the components table.
type PostgresComponentDAO struct {
	db *sql.DB
}

// NewPostgresComponentDAO creates a new PostgresComponentDAO instance.
func NewPostgresComponentDAO(db *sql.DB) *PostgresComponentDAO {
	return &PostgresComponentDAO{db: db}
}

// getExecutor returns either the transaction or the database connection.
func (c *PostgresComponentDAO) getExecutor(tx *sql.Tx) interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
} {
	if tx != nil {
		return tx
	}
	return c.db
}

// CreateComponent creates a new component instance within a transaction.
func (c *PostgresComponentDAO) CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error) {
	executor := c.getExecutor(tx)
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
			component_status, min_threshold, max_threshold, max_running_hours,
			device_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING component_id, component_name, component_type, component_subtype,
				  component_status, min_threshold, max_threshold, max_running_hours, device_id
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
		component.DeviceID,
	).Scan(
		&component.ComponentID,
		&component.ComponentName,
		&component.ComponentType,
		&component.ComponentSubtype,
		&component.ComponentStatus,
		&minThreshold,
		&maxThreshold,
		&maxRunningHours,
		&component.DeviceID,
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
func (c *PostgresComponentDAO) GetComponentByID(id string) (*models.Component, error) {
	stmt := `
		SELECT component_id, component_name, component_type, component_subtype,
			   component_status, min_threshold, max_threshold, max_running_hours, current_running_hours, device_id
		FROM components
		WHERE component_id = $1
	`
	row := c.db.QueryRow(stmt, id)
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
		&component.CurrentRunningHours,
		&component.DeviceID,
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

// UpdateComponentStatus updates the status of a component within a transaction.
func (c *PostgresComponentDAO) UpdateComponentStatus(tx *sql.Tx, id string, status string) error {
	executor := c.getExecutor(tx)
	_, err := executor.Exec("UPDATE components SET component_status = $1 WHERE component_id = $2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update component status for %s: %w", id, err)
	}
	return nil
}

// GetComponentsByDeviceID retrieves all components linked to a specific device.
func (c *PostgresComponentDAO) GetComponentsByDeviceID(id string) ([]*models.Component, error) {
	rows, err := c.db.Query(`
		SELECT c.component_id, c.component_name, c.component_type, c.component_subtype,
			c.component_status, c.min_threshold, c.max_threshold, c.max_running_hours,
			c.current_running_hours
		FROM public.components c
		WHERE c.device_id = $1
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
func (c *PostgresComponentDAO) UpdateComponentRange(tx *sql.Tx, component *models.Component) error {
	executor := c.getExecutor(tx)
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

// UpdateComponentRunningHours updates the running hours of a component within a transaction.
func (c *PostgresComponentDAO) UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error {
	executor := c.getExecutor(tx)
	_, err := executor.Exec("UPDATE components SET current_running_hours = $1 WHERE component_id = $2", hours, id)
	if err != nil {
		return fmt.Errorf("failed to update component running hours for %s: %w", id, err)
	}
	return nil
}

// GetSensorsByDeviceID retrieves all sensor components linked to a specific device.
func (c *PostgresComponentDAO) GetSensorsByDeviceID(id string) ([]*models.Component, error) {
	rows, err := c.db.Query(`
		SELECT c.component_id, c.component_name, c.component_type, c.component_subtype,
			c.component_status, c.min_threshold, c.max_threshold, c.max_running_hours,
			c.current_running_hours
		FROM public.components c
		WHERE c.device_id = $1 AND c.component_type = 'sensor'
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

// ResetComponentRunningHours resets the running hours of a component within a transaction.
func (c *PostgresComponentDAO) ResetComponentRunningHours(tx *sql.Tx, id string) error {
	executor := c.getExecutor(tx)
	_, err := executor.Exec("UPDATE components SET current_running_hours = 0 WHERE component_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to reset component running hours for %s: %w", id, err)
	}
	return nil
}

// UpdateComponentConfig updates the configuration of a component within a transaction.
func (c *PostgresComponentDAO) UpdateComponentConfig(tx *sql.Tx, component models.ComponentConfig) error {
	executor := c.getExecutor(tx)
	var minThreshold, maxThreshold sql.NullFloat64
	if component.MinThreshold != nil {
		minThreshold = sql.NullFloat64{Float64: *component.MinThreshold, Valid: true}
	}
	if component.MaxThreshold != nil {
		maxThreshold = sql.NullFloat64{Float64: *component.MaxThreshold, Valid: true}
	}
	var maxRunningHours sql.NullInt32
	if component.MaxRunningHours != nil {
		maxRunningHours = sql.NullInt32{Int32: *component.MaxRunningHours, Valid: true}
	}

	_, err := executor.Exec(
		`UPDATE components SET min_threshold = $1, max_threshold = $2, max_running_hours = $3 WHERE component_id = $4`,
		minThreshold,
		maxThreshold,
		maxRunningHours,
		component.ComponentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update component config for %s: %w", component.ComponentID, err)
	}
	return nil
}

// UserHasAccessToComponent checks if a user has access to a specific component.
func (c *PostgresComponentDAO) UserHasAccessToComponent(userID int, componentID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM components AS c
			JOIN devices AS d ON c.device_id = d.device_id
			JOIN device_location AS dl ON d.device_id = dl.device_id AND dl.is_current = TRUE
			JOIN locations AS l ON dl.location_id = l.location_id
			JOIN user_site AS us ON l.site_id = us.site_id
			WHERE us.user_id = $1 AND c.component_id = $2
		)
	`
	var hasAccess bool
	err := c.db.QueryRow(query, userID, componentID).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check user access to component: %w", err)
	}
	return hasAccess, nil
}

// LinkComponentToDevice links an existing component to a device within a transaction.
func (c *PostgresComponentDAO) LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error {
	executor := c.getExecutor(tx)
	query := `
		UPDATE public.components
		SET device_id = $1
		WHERE component_id = $2
	`
	_, err := executor.Exec(query, deviceID, componentID)
	if err != nil {
		return fmt.Errorf("failed to link component %s to device %s: %w", componentID, deviceID, err)
	}
	return nil
}

// FindAllWithComponentsAndLocations retrieves devices with their components, locations, and site names.
func (c *PostgresComponentDAO) FindAllWithComponentsAndLocations(
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

	rows, err := c.db.QueryContext(ctx, query, args...)
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
			components, err := c.GetComponentsByDeviceID(deviceInfo.DeviceWithComponents.Device.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get components for device %s: %w", deviceInfo.DeviceWithComponents.Device.DeviceID, err)
			}
			deviceInfo.DeviceWithComponents.Components = components
		}
	}

	return devices, nil
}

// FindAllWithComponentsAndLocations retrieves devices with their components, locations, and site names.
func (c *PostgresDeviceDAO) FindAllWithComponentsAndLocations(
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

	if search != "" {
		query += fmt.Sprintf(`
			WHERE LOWER(d.device_id) LIKE LOWER('%%' || $%d || '%%')
			OR LOWER(l.location_name) LIKE LOWER('%%' || $%d || '%%')
			OR LOWER(s.site_name) LIKE LOWER('%%' || $%d || '%%')
		`, argIndex, argIndex, argIndex)
		args = append(args, search)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY d.device_id LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := c.db.QueryContext(ctx, query, args...)
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

	for _, deviceInfo := range devices {
		if deviceInfo.DeviceWithComponents != nil && deviceInfo.DeviceWithComponents.Device != nil {
			components, err := c.GetComponentsByDeviceID(deviceInfo.DeviceWithComponents.Device.DeviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get components for device %s: %w", deviceInfo.DeviceWithComponents.Device.DeviceID, err)
			}
			deviceInfo.DeviceWithComponents.Components = components
		}
	}

	return devices, nil
}

// GetCOmponentsBYDeviceID retrieves all components linked to a specific device.
func (c *PostgresDeviceDAO) GetComponentsByDeviceID(id string) ([]*models.Component, error) {
	rows, err := c.db.Query(`
		SELECT c.component_id, c.component_name, c.component_type, c.component_subtype,
			c.component_status, c.min_threshold, c.max_threshold, c.max_running_hours,
			c.current_running_hours
		FROM public.components c
		WHERE c.device_id = $1
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

// HandleDeviceAlert inserts a new alert log for a component within a transaction.
func (c *PostgresComponentDAO) HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error {
	executor := c.getExecutor(tx)
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
func (c *PostgresComponentDAO) GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error) {
	rows, err := c.db.Query("SELECT log_id, component_id, log_timestamp, log_content, log_read FROM component_log WHERE component_id = $1", id)
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
func (c *PostgresComponentDAO) GetcomponentLogsByDeviceIDAndComponentID(deviceID string, componentID string) ([]*models.ComponentLog, error) {
	rows, err := c.db.Query(`
        SELECT cl.log_id, cl.component_id, cl.log_timestamp, cl.log_content, cl.log_read
        FROM component_log cl
        JOIN components comp ON cl.component_id = comp.component_id
        WHERE comp.device_id = $1 AND cl.component_id = $2
    `, deviceID, componentID)
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
func (c *PostgresComponentDAO) GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error) {
	rows, err := c.db.Query(`
        SELECT cl.log_id, cl.component_id, cl.log_timestamp, cl.log_content, cl.log_read
        FROM component_log cl
        JOIN components comp ON cl.component_id = comp.component_id
        WHERE comp.device_id = $1
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
func (c *PostgresComponentDAO) GetAllLogsByUser(userID int) ([]*models.ComponentLog, error) {
	rows, err := c.db.Query(`
        SELECT cl.log_id, cl.component_id, cl.log_timestamp, cl.log_content, cl.log_read
        FROM component_log cl
        JOIN components comp ON cl.component_id = comp.component_id
        JOIN device_location dl ON comp.device_id = dl.device_id AND dl.is_current = true
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

// MarkComponentLogsAsRead marks component logs as read for a specific component within a transaction.
func (c *PostgresComponentDAO) MarkComponentLogsAsRead(tx *sql.Tx, componentID string, logIds []int) error {
	if len(logIds) == 0 {
		return nil // No logs to mark as read
	}

	executor := c.getExecutor(tx)
	placeholders := make([]string, len(logIds))
	args := make([]interface{}, len(logIds)+1)
	for i, id := range logIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(logIds)] = componentID // Last argument is the component ID

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
func (c *PostgresComponentDAO) MarkAllLogsAsRead(tx *sql.Tx, userID int) error {
	executor := c.getExecutor(tx)
	query := `
       UPDATE component_log AS cl
       SET log_read = TRUE
       FROM components AS comp
       JOIN device_location AS dl ON comp.device_id = dl.device_id AND dl.is_current = TRUE
       JOIN locations AS l ON dl.location_id = l.location_id
       JOIN user_site AS us ON l.site_id = us.site_id
       WHERE us.user_id = $1 AND cl.component_id = comp.component_id
    `
	_, err := executor.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all logs as read for user %d: %w", userID, err)
	}
	return nil
}
