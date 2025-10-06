package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// PostgresLocationRepository implements LocationDAO using PostgreSQL.
type PostgresLocationRepository struct {
	db *sql.DB
}

// NewPostgresLocationRepository creates a new PostgresLocationRepository.
func NewPostgresLocationRepository(db *sql.DB) *PostgresLocationRepository {
	return &PostgresLocationRepository{db: db}
}

// --- Location Operations ---

// InsertLocation inserts a location into the database.
func (r *PostgresLocationRepository) InsertLocation(location models.Location) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO locations (location_name, location_description, site_id) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, location.Name, location.Description, location.SiteID); err != nil {
		log.Printf("Error inserting location into database: %v\n", err)
		return err
	}
	log.Printf("Location %s inserted into database.\n", *location.Name)
	return nil
}

// LocationExists checks if a location exists in the database.
func (r *PostgresLocationRepository) LocationExists(ID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `SELECT 1 FROM locations WHERE location_id = $1 LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, ID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Printf("Error checking location existence: %v\n", err)
		return false, err
	}
	return true, nil
}

// GetAllLocations retrieves all locations from the database with pagination and search.
func (r *PostgresLocationRepository) GetAllLocations(ctx context.Context, page int, limit int, term string) ([]*models.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
       SELECT location_id, location_name, location_description, site_id
       FROM locations
       WHERE location_name ILIKE '%' || $1 || '%'
       ORDER BY location_name
       LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, term, limit, (page-1)*limit)
	if err != nil {
		log.Printf("Error getting all locations: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var locations []*models.Location
	for rows.Next() {
		var location models.Location
		var siteID sql.NullInt32
		if err := rows.Scan(&location.ID, &location.Name, &location.Description, &siteID); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return nil, err
		}
		if siteID.Valid {
			id := int(siteID.Int32)
			location.SiteID = &id
		} else {
			location.SiteID = nil
		}
		locations = append(locations, &location)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location rows: %v\n", err)
		return nil, err
	}
	return locations, nil
}

// GetLocationByID retrieves a location by its ID.
func (r *PostgresLocationRepository) GetLocationByID(id string) (models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT location_id, location_name, location_description FROM locations WHERE location_id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var location models.Location
	if err := row.Scan(&location.ID, &location.Name, &location.Description); err != nil {
		if err == sql.ErrNoRows {
			return models.Location{}, nil
		}
		log.Printf("Error scanning location row: %v\n", err)
		return models.Location{}, err
	}
	return location, nil
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

// DeleteLocation deletes a location by its ID.
func (r *PostgresLocationRepository) DeleteLocation(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM locations WHERE location_id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		log.Printf("Error deleting location with ID %s: %v\n", id, err)
		return err
	}
	log.Printf("Location with ID %s deleted from database.\n", id)
	return nil
}

// DeleteSite deletes a site by its ID.
func (r *PostgresLocationRepository) DeleteSite(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `DELETE FROM sites WHERE site_id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		log.Printf("Error deleting site with ID %s: %v\n", id, err)
		return err
	}
	log.Printf("Site with ID %s deleted from database.\n", id)
	return nil
}

// ModifyLocation modifies a location in the database.
func (r *PostgresLocationRepository) ModifyLocation(ctx context.Context, location models.Location) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE locations SET location_name = $1, location_description = $2 WHERE location_id = $3`
	if _, err := r.db.ExecContext(ctx, query, location.Name, location.Description, location.ID); err != nil {
		log.Printf("Error modifying location with ID %s: %v\n", location.ID, err)
		return err
	}
	log.Printf("Location with ID %s modified in database.\n", location.ID)
	return nil
}

// --- Site Operations ---

// CreateSite creates a new site in the database.
func (r *PostgresLocationRepository) CreateSite(ctx context.Context, site models.Site) (*models.Site, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO sites (site_name, site_address) VALUES ($1, $2)`
	if _, err := r.db.ExecContext(ctx, query, site.Name, site.Address); err != nil {
		log.Printf("Error inserting site into database: %v\n", err)
		return nil, err
	}
	log.Printf("Site %s inserted into database.\n", *site.Name)
	return &site, nil
}

// GetSitesWithPagination retrieves sites with pagination and search.
func (r *PostgresLocationRepository) GetSitesWithPagination(ctx context.Context, page int, limit int, term string) ([]*models.Site, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT site_id, site_name, site_address FROM sites WHERE site_name ILIKE '%' || $1 || '%' ORDER BY site_name LIMIT $2 OFFSET $3`
	offset := (page - 1) * limit
	rows, err := r.db.QueryContext(ctx, query, term, limit, offset)
	if err != nil {
		log.Printf("Error querying sites with pagination: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var sites []*models.Site
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.ID, &site.Name, &site.Address); err != nil {
			log.Printf("Error scanning site row: %v\n", err)
			return nil, err
		}
		sites = append(sites, &site)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating site rows: %v\n", err)
		return nil, err
	}
	return sites, nil
}

// CountSites counts the total number of sites.
func (r *PostgresLocationRepository) CountSites(ctx context.Context, term string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM sites WHERE site_name ILIKE '%' || $1 || '%'`
	var count int
	if err := r.db.QueryRowContext(ctx, query, term).Scan(&count); err != nil {
		log.Printf("Error counting sites: %v\n", err)
		return 0, err
	}
	return count, nil
}

// --- Combined/Related Operations ---

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
			deviceID            sql.NullString
			deviceStatus        sql.NullString
			deviceLastSeen      sql.NullTime
			userID              sql.NullInt64
			username            sql.NullString
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
					ID:          locationID,
					Name:        locationName,
					Description: locationDescription,
				},
				Devices: []*models.DeviceWithComponents{},
				Users:   []*models.User{},
			}
		}

		location := locationsMap[*locationID]

		if deviceID.Valid {
			device := models.Device{
				DeviceID: deviceID.String,
				Status:   deviceStatus.String,
				LastSeen: deviceLastSeen.Time,
			}
			location.Devices = append(location.Devices, &models.DeviceWithComponents{Device: &device})
		}

		if userID.Valid {
			user := models.User{
				ID:   int(userID.Int64),
				Name: &username.String,
			}
			location.Users = append(location.Users, &user)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location with devices and users rows: %v\n", err)
		return nil, err
	}

	locationsSlice := make([]*models.LocationWithUsersDevices, 0, len(locationsMap))
	for _, loc := range locationsMap {
		locationsSlice = append(locationsSlice, loc)
	}

	return locationsSlice, nil
}

// GetMySites retrieves sites associated with a user ID.
func (r *PostgresLocationRepository) GetMySites(ctx context.Context, userID int) ([]models.Site, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT s.site_id, s.site_name, s.site_address
	FROM sites s
	JOIN locations l ON s.site_id = l.site_id
	JOIN device_location dl ON l.location_id = dl.location_id
	JOIN users u ON dl.id = u.id
	WHERE u.id = $1
	GROUP BY s.site_id`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error getting sites for user ID %d: %v\n", userID, err)
		return []models.Site{}, err
	}
	defer rows.Close()

	sites := []models.Site{}
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.ID, &site.Name, &site.Address); err != nil {
			log.Printf("Error scanning site row: %v\n", err)
			return []models.Site{}, err
		}
		sites = append(sites, site)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating site rows: %v\n", err)
		return []models.Site{}, err
	}
	return sites, nil
}

// GetComponentsBylocationID retrieves components associated with a location ID.
func (r *PostgresLocationRepository) GetComponentsBylocationID(locationID string) ([]models.Component, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT
       c.component_id,
       c.component_type
    FROM components c
    JOIN device_components dc ON c.component_id = dc.component_id
    JOIN device_location dl ON dc.device_id = dl.device_id
    WHERE dl.location_id = $1 `

	rows, err := r.db.QueryContext(ctx, query, locationID)
	if err != nil {
		log.Printf("Error getting components by location ID: %v\n", err)
		return []models.Component{}, err
	}
	defer rows.Close()

	components := []models.Component{}
	for rows.Next() {
		var component models.Component
		if err := rows.Scan(&component.ComponentID, &component.ComponentID); err != nil {
			log.Printf("Error scanning component row: %v\n", err)
			return []models.Component{}, err
		}
		components = append(components, component)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating component rows: %v\n", err)
		return []models.Component{}, err
	}
	return components, nil
}

// GetDevicesBylocationID retrieves devices associated with a location ID.
func (r *PostgresLocationRepository) GetDevicesBylocationID(id string) ([]models.Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT d.device_id, d.status, d.last_seen, d.power, d.voltage, d.current
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
		if err := rows.Scan(&device.DeviceID, &device.Status, &device.LastSeen, &device.Power, &device.Voltage, &device.Current); err != nil {
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

// CheckUserAccessToLocation checks if a user has access to a location.
func (r *PostgresLocationRepository) CheckUserAccessToLocation(userID int, locationID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT 1
	FROM users u
	JOIN user_site us ON u.id = us.user_id
	JOIN locations l ON us.site_id = l.site_id
	WHERE u.id = $1 AND l.location_id = $2
	LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, userID, locationID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		log.Printf("Error checking user access to location: %v\n", err)
		return false, err
	}
	return true, nil
}

// CheckUserAccessToSite checks if a user has access to a site.
func (r *PostgresLocationRepository) CheckUserAccessToSite(userID int, siteID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT 1
	          FROM user_site
	          WHERE user_id = $1 AND site_id = $2
	          LIMIT 1`

	var exists int
	err := r.db.QueryRowContext(ctx, query, userID, siteID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Printf("Error checking user access to site: %v\n", err)
		return false, err
	}
	return true, nil
}
func (r *PostgresLocationRepository) GetLocationsBySiteIDs(ctx context.Context, siteIDs []int, page int, limit int, term string) ([]*models.Location, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Parse the siteID string to an integer

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Create a dynamic placeholder string for the IN clause
	placeholders := ""
	args := make([]interface{}, len(siteIDs)+3) // +3 for term, limit, offset
	for i, id := range siteIDs {
		placeholders += fmt.Sprintf("$%d,", i+1)
		args[i] = id
	}
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	// Add term, limit, and offset to args
	args[len(siteIDs)] = term
	args[len(siteIDs)+1] = limit
	args[len(siteIDs)+2] = offset

	query := fmt.Sprintf(`
		SELECT location_id, location_name, location_description, site_id
		FROM locations
		WHERE site_id IN (%s) AND location_name ILIKE '%%' || $%d || '%%'
		ORDER BY location_name
		LIMIT $%d OFFSET $%d
	`, placeholders, len(siteIDs)+1, len(siteIDs)+2, len(siteIDs)+3)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error querying locations by site IDs: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var locations []*models.Location
	for rows.Next() {
		var location models.Location
		var siteID sql.NullInt32
		if err := rows.Scan(&location.ID, &location.Name, &location.Description, &siteID); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return nil, err
		}
		if siteID.Valid {
			id := int(siteID.Int32)
			location.SiteID = &id
		} else {
			location.SiteID = nil
		}
		locations = append(locations, &location)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location rows: %v\n", err)
		return nil, err
	}
	return locations, nil
}
