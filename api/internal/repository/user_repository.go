package repository

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"log"
	"time"
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
func (r *PostgresUserRepository) CreateUser(auth0ID string, auth0_email string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
func (r *PostgresUserRepository) UserExists(auth0_id string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
func (r *PostgresUserRepository) FindUserByID(ID int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, ID)

	var user struct {
		ID    int
		Name  *string // Name can be NULL, so use a pointer
		Email string
		Role  sql.NullString
	}
	if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by ID: %v\n", err)
		return nil, err
	}

	return &models.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

// UpdateUser updates a user in the database.
func (r *PostgresUserRepository) UpdateUser(user models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
		return nil, err
	}
	log.Printf("User %s updated in database.\n", user.ID)
	return &user, nil
}

// DeleteUser deletes a user from the database.
func (r *PostgresUserRepository) DeleteUser(ID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
func (r *PostgresUserRepository) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by email: %v\n", err)
		return nil, err
	}

	return &user, nil
}
func (r *PostgresUserRepository) GetAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error getting all users: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var list []models.User
	for rows.Next() {
		var user models.User
		var name sql.NullString
		if err := rows.Scan(&user.ID, &name, &user.Email); err != nil {
			log.Printf("Error scanning user row: %v\n", err)
			return nil, err
		}
		user.Name = &name.String
		list = append(list, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating user rows: %v\n", err)
		return nil, err
	}

	return list, nil
}

// AsignUser assigns a user to a location in the database.
func (r *PostgresUserRepository) AsignUser(userID, locationID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO user_location (user_id, location_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, userID, locationID)
	if err != nil {
		log.Printf("Error assigning user to location: %v\n", err)
		return err
	}
	log.Printf("User %d assigned to location %d.\n", userID, locationID)
	return nil
}

// GetUserLocations retrieves all locations assigned to a user.
func (r *PostgresUserRepository) GetUserLocations(userID int) ([]models.Location, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT l.location_id, l.location_name, l.location_description FROM locations l JOIN user_location ul ON l.location_id = ul.location_id WHERE ul.user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error getting user locations: %v\n", err)
		return []models.Location{}, err // Return empty slice on error
	}
	defer rows.Close()
	var locations []models.Location
	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Description); err != nil {
			log.Printf("Error scanning location row: %v\n", err)
			return []models.Location{}, err // Return empty slice on scan error
		}
		locations = append(locations, location)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating location rows: %v\n", err)
		return []models.Location{}, err // Return empty slice on iteration error
	}
	log.Printf("User %d has %d locations.\n", userID, len(locations))
	return locations, nil // Return the (potentially empty) locations slice
}

// GetUsersLocations retrieves all locations assigned to a user.
func (r *PostgresUserRepository) FindAllWithLocations(ctx context.Context, limit int, offset int, search string) ([]*models.UserLocations, error) {
	query := `
		SELECT
			u.id, u.name, u.email, u.created_at, u.auth0_id,
			l.location_id, l.location_name, l.location_description
			FROM users u
			LEFT JOIN user_location ul ON u.id = ul.user_id
			LEFT JOIN locations l ON ul.location_id = l.location_id
			WHERE u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%'
			LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, search, limit, offset)
	if err != nil {
		log.Printf("Error getting users and their locations: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	usersLocationMap := make(map[int]*models.UserLocations)

	for rows.Next() {
		var user models.User
		var locationID *int
		var locationName, locationDescription *string

		if err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.Auth0ID,
			&locationID, &locationName, &locationDescription,
		); err != nil {
			log.Printf("Error scanning user location row: %v\n", err)
			return nil, err
		}
		// If user.name is NULL, set it to an empty string
		userLocation, ok := usersLocationMap[user.ID]
		if !ok {
			userLocation = &models.UserLocations{
				User:      user,
				Locations: []*models.Location{},
			}
			usersLocationMap[user.ID] = userLocation
		}

		if locationID != nil {
			location := &models.Location{
				ID:          locationID,
				Name:        nil,
				Description: nil,
			}
			if locationName != nil {
				location.Name = locationName
			}
			if locationDescription != nil {
				location.Description = locationDescription
			}
			userLocation.Locations = append(userLocation.Locations, location)
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating user location rows: %v\n", err)
		return nil, err
	}

	var usersLocations []*models.UserLocations
	for _, ul := range usersLocationMap {
		usersLocations = append(usersLocations, ul)
	}

	return usersLocations, nil
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

// UpdateUserLocation updates the user's location in the database.
func (r *PostgresUserRepository) UpdateUserLocation(userID int, locationIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil) // Start a transaction
	if err != nil {
		log.Printf("Error starting transaction for user %d location update: %v\n", userID, err)
		return err
	}
	defer tx.Rollback() // Rollback on error, if Commit is not called

	// 1. Delete all existing locations for the user
	deleteQuery := `DELETE FROM user_location WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, userID)
	if err != nil {
		log.Printf("Error deleting existing locations for user %d: %v\n", userID, err)
		return err
	}
	log.Printf("Deleted existing locations for user %d.\n", userID) // Log deletion

	// 2. Insert new locations for each ID in the slice
	insertQuery := `INSERT INTO user_location (user_id, location_id) VALUES ($1, $2)`
	for _, locationID := range locationIDs {
		_, err := tx.ExecContext(ctx, insertQuery, userID, locationID)
		if err != nil {
			log.Printf("Error inserting location %d for user %d: %v\n", locationID, userID, err)
			return err // Rollback will be called by defer
		}
	}
	log.Printf("Inserted new locations for user %d: %v.\n", userID, locationIDs) // Log insertion

	if err := tx.Commit(); err != nil { // Commit the transaction
		log.Printf("Error committing transaction for user %d location update: %v\n", userID, err)
		return err
	}

	log.Printf("User %d updated with new locations %v.\n", userID, locationIDs)
	return nil
}

// UpdateUserName updates the user's name in the database.
func (r *PostgresUserRepository) UpdateUserName(id int, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET name = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		log.Printf("Error updating user name: %v\n", err)
		return err
	}
	log.Printf("User %d updated with new name %s.\n", id, name)
	return nil
}
