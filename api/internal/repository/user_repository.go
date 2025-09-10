package repository

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"log"
)

// PostgresUserRepository implements UserDAO using PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository.
func NewPostgresUserRepository(db *sql.DB) dao.UserDAO {
	return &PostgresUserRepository{db: db}
}

// CreateUser inserts a user with the Auth0 ID and email into the database and returns the user ID.
func (r *PostgresUserRepository) CreateUser(ctx context.Context, auth0ID string, auth0_email string) (int, error) {
	var userID int
	query := `INSERT INTO users (auth0_id, email) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, auth0ID, auth0_email).Scan(&userID)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		return 0, err
	}
	log.Printf("User with Auth0 ID %s inserted into database with ID %d.\n", auth0ID, userID)
	return userID, nil
}

// UserExists checks if a user exists in the database.
func (r *PostgresUserRepository) UserExists(ctx context.Context, auth0_id string) (int, error) {
	const query = `SELECT id FROM users WHERE auth0_id = $1 LIMIT 1`
	var userID int
	err := r.db.QueryRowContext(ctx, query, auth0_id).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		log.Printf("Error checking user existence: %v\n", err)
		return 0, err
	}
	return userID, nil
}

// FindUserByID retrieves a user by their ID from the database.
func (r *PostgresUserRepository) FindUserByID(ctx context.Context, ID int) (*models.User, error) {
	query := `SELECT id, name, email FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, ID)

	var user models.User
	var name sql.NullString // Use sql.NullString for nullable columns
	if err := row.Scan(&user.ID, &name, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by ID: %v\n", err)
		return nil, err
	}
	if name.Valid {
		user.Name = &name.String
	} else {
		user.Name = nil
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetUsernameByAuth0ID(ctx context.Context, id string) (string, error) {
	query := `SELECT name FROM users WHERE auth0_id = $1`
	var name sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		log.Printf("Error fetching username by Auth0 ID: %v\n", err)
		return "", err
	}
	if name.Valid {
		return name.String, nil
	}
	return "", nil
}

// UpdateUser updates a user in the database.
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user models.User) (*models.User, error) {
	query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
		return nil, err
	}
	log.Printf("User %d updated in database.\n", user.ID)
	return &user, nil
}

// DeleteUser deletes a user from the database.
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, ID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, ID)
	if err != nil {
		log.Printf("Error deleting user from database: %v\n", err)
		return err
	}
	log.Printf("User %d deleted from database.\n", ID)
	return nil
}

// GetUserByEmail retrieves a user by their email from the database.
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	var user models.User
	var name sql.NullString
	if err := row.Scan(&user.ID, &name, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by email: %v\n", err)
		return nil, err
	}
	if name.Valid {
		user.Name = &name.String
	} else {
		user.Name = nil
	}

	return &user, nil
}
func (r *PostgresUserRepository) GetUsers(ctx context.Context, limit int, offset int, term string) ([]*models.User, error) {
	query := `SELECT id, name, email FROM users WHERE name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, term, limit, offset)
	if err != nil {
		log.Printf("Error querying users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var name sql.NullString
		if err := rows.Scan(&user.ID, &name, &user.Email); err != nil {
			log.Printf("Error scanning user row: %v\n", err)
			return nil, err
		}
		if name.Valid {
			user.Name = &name.String
		} else {
			user.Name = nil
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating user rows: %v\n", err)
		return nil, err
	}

	return users, nil
}

// AsignUser assigns a user to a site in the database.
func (r *PostgresUserRepository) AsignUser(ctx context.Context, userID, siteID int) error {
	query := `INSERT INTO user_site (user_id, site_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, userID, siteID)
	if err != nil {
		log.Printf("Error assigning user to site: %v\n", err)
		return err
	}
	log.Printf("User %d assigned to site %d.\n", userID, siteID)
	return nil
}

// GetUserLocations retrieves all locations assigned to a user.
func (r *PostgresUserRepository) GetUserLocations(ctx context.Context, userID int) ([]models.Location, error) {
	query := `
		SELECT DISTINCT
			l.location_id,
			l.location_name,
			l.location_description
		FROM locations AS l
		JOIN sites AS s ON l.site_id = s.site_id
		JOIN user_site AS us ON s.site_id = us.site_id
		WHERE us.user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error getting user locations: %v\n", err)
		return []models.Location{}, err
	}
	defer rows.Close()
	var locations []models.Location
	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Description); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return []models.Location{}, err
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location rows: %v\n", err)
		return []models.Location{}, err
	}
	log.Printf("User %d has %d locations.\n", userID, len(locations))
	return locations, nil
}

// GetUserSites retrieves all sites assigned to a user.
func (r *PostgresUserRepository) GetUserSites(ctx context.Context, userID int) ([]models.Site, error) {
	query := `
		SELECT s.site_id, s.site_name, s.site_address
		FROM sites AS s
		JOIN user_site AS us ON s.site_id = us.site_id
		WHERE us.user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error querying user sites: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.ID, &site.Name, &site.Address); err != nil {
			log.Printf("Error scanning site row: %v\n", err)
			return nil, err
		}
		sites = append(sites, site)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating site rows: %v\n", err)
		return nil, err
	}

	return sites, nil
}

// CountAll counts all users in the database with optional search.
func (r *PostgresUserRepository) CountAll(ctx context.Context, search string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'`
	row := r.db.QueryRowContext(ctx, query, search)
	var count int
	if err := row.Scan(&count); err != nil {
		log.Printf("Error counting users: %v\n", err)
		return 0, err
	}
	return count, nil
}

// UpdateUserSites updates the user's assigned sites in the database.
func (r *PostgresUserRepository) UpdateUserSites(ctx context.Context, userID int, siteIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction for user %d site update: %v\n", userID, err)
		return err
	}
	defer tx.Rollback()

	// 1. Delete all existing site assignments for the user
	deleteQuery := `DELETE FROM user_site WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, userID)
	if err != nil {
		log.Printf("Error deleting existing sites for user %d: %v\n", userID, err)
		return err
	}
	log.Printf("Deleted existing sites for user %d.\n", userID)

	// 2. Insert new site assignments for each ID in the slice
	insertQuery := `INSERT INTO user_site (user_id, site_id) VALUES ($1, $2)`
	for _, siteID := range siteIDs {
		_, err := tx.ExecContext(ctx, insertQuery, userID, siteID)
		if err != nil {
			log.Printf("Error inserting site %d for user %d: %v\n", siteID, userID, err)
			return err
		}
	}
	log.Printf("Inserted new sites for user %d: %v.\n", userID, siteIDs)

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction for user %d site update: %v\n", userID, err)
		return err
	}

	log.Printf("User %d updated with new sites %v.\n", userID, siteIDs)
	return nil
}

// UpdateUserName updates the user's name in the database.
func (r *PostgresUserRepository) UpdateUserName(ctx context.Context, id int, name string) error {
	query := `UPDATE users SET name = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		log.Printf("Error updating user name: %v\n", err)
		return err
	}
	log.Printf("User %d updated with new name %s.\n", id, name)
	return nil
}
