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
func (r *PostgresLocationRepository) GetAllLocations(ctx context.Context, page int, limit int, term string) ([]*models.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT location_id, location_name, location_description
	FROM locations
	WHERE location_name ILIKE '%' || $1 || '%'
	ORDER BY location_name
	LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, term, limit, (page-1)*limit)
	if err != nil {
		log.Printf("Error getting all locations: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	locations := []*models.Location{}

	for rows.Next() {
		var location models.Location

		if err := rows.Scan(
			&location.ID,
			&location.Name,
			&location.Description,
		); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return nil, err
		}
		locations = append(locations, &location)
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

// GetsensorsByLocationID retrieves sensors associated with a location ID.
func (r *PostgresLocationRepository) GetsensorsByLocationID(locationID string) ([]models.Sensor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT
       c.sensor_id,
       c.sensor_type
    FROM sensors c
    JOIN device_sensors dc ON c.sensor_id = dc.sensor_id
    JOIN device_location dl ON dc.device_id = dl.device_id
    WHERE dl.location_id = $1 `

	rows, err := r.db.QueryContext(ctx, query, locationID)
	if err != nil {
		log.Printf("Error getting sensors by location ID: %v\n", err)
		return []models.Sensor{}, err
	}
	defer rows.Close()

	sensors := []models.Sensor{} // Initialize as an empty slice

	for rows.Next() {
		var sensor models.Sensor

		if err := rows.Scan(
			&sensor.SensorID,
			&sensor.SensorType,
		); err != nil {
			log.Printf("Error scanning sensor row: %v\n", err)
			return []models.Sensor{}, err
		}
		sensors = append(sensors, sensor)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating sensor rows: %v\n", err)
		return []models.Sensor{}, err
	}

	return sensors, nil
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

// GetLocationsDevicesUsers retrieves locations with associated devices and users.
func (r *PostgresLocationRepository) GetLocationsDevicesUsers(ctx context.Context, page int, limit int, term string) ([]*models.LocationWithUsersDevices, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT l.location_id, l.location_name, l.location_description,
       d.device_id, d.status, d.last_seen,
       u.id, u.name
    FROM locations l
    LEFT JOIN device_location dl ON l.location_id = dl.location_id
    LEFT JOIN devices d ON dl.device_id = d.device_id
    LEFT JOIN users u ON dl.id = u.id -- Corrected JOIN condition
    WHERE l.location_name ILIKE '%' || $1 || '%'
    ORDER BY l.location_name
    LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, term, limit, (page-1)*limit)
	if err != nil {
		log.Printf("Error getting locations with devices and users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	locationsMap := make(map[int]*models.LocationWithUsersDevices)

	for rows.Next() {
		var (
			locationID          *int
			locationName        *string
			locationDescription *string
			deviceID            *string
			deviceStatus        *string
			deviceLastSeen      *time.Time
			userID              *int
			username            *string
		)

		if err := rows.Scan(
			&locationID,
			&locationName,
			&locationDescription,
			&deviceID,
			&deviceStatus,
			&deviceLastSeen,
			&userID,
			&username,
		); err != nil {
			log.Printf("Error scanning location with devices and users row: %v\n", err)
			return nil, err
		}

		if _, ok := locationsMap[*locationID]; !ok {
			locationsMap[*locationID] = &models.LocationWithUsersDevices{
				Location: models.Location{
					ID:          locationID,          // Assign pointer to int
					Name:        locationName,        // Assign pointer to string
					Description: locationDescription, // Assign pointer to string
				},
				Devices: []*models.DeviceWithsensors{}, // Initialize with the correct type
				Users:   []*models.User{},
			}
		}

		location := locationsMap[*locationID]

		if deviceID != nil {
			location.Devices = append(location.Devices, &models.DeviceWithsensors{ // Use the correct struct
				Device: &models.Device{
					DeviceID: *deviceID,       // Assign pointer to int
					Status:   *deviceStatus,   // Assign pointer to string
					LastSeen: *deviceLastSeen, // Assign pointer to time.Time
				},
				// sensors field will be nil as it's not fetched in this query
			})
		}

		if userID != nil {
			location.Users = append(location.Users, &models.User{
				ID:   *userID,
				Name: username,
			})
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location with devices and users rows: %v\n", err)
		return nil, err
	}

	var locationsSlice []*models.LocationWithUsersDevices
	for _, loc := range locationsMap {
		locationsSlice = append(locationsSlice, loc)
	}

	return locationsSlice, nil
}

// CountAll counts all locations in the database.
func (r *PostgresLocationRepository) CountAll(ctx context.Context, term string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM locations WHERE location_name ILIKE '%' || $1 || '%'`
	var count int
	err := r.db.QueryRowContext(ctx, query, term).Scan(&count)
	if err != nil {
		log.Printf("Error counting locations: %v\n", err)
		return 0, err
	}

	return count, nil
}

// DeleteLocation deletes a location by its ID from the database.
func (r *PostgresLocationRepository) DeleteLocation(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM locations WHERE location_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error deleting location with ID %s: %v\n", id, err)
		return err
	}

	log.Printf("Location with ID %s deleted from database.\n", id)
	return nil
}

// ModifyLocation modifies a location in the database.
func (r *PostgresLocationRepository) ModifyLocation(ctx context.Context, location models.Location) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE locations SET location_name = $1, location_description = $2 WHERE location_id = $3`
	_, err := r.db.ExecContext(ctx, query, location.Name, location.Description, location.ID)
	if err != nil {
		log.Printf("Error modifying location with ID %s: %v\n", location.ID, err)
		return err
	}

	log.Printf("Location with ID %s modified in database.\n", location.ID)
	return nil
}
