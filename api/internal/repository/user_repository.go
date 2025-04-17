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

	query := `SELECT id, name, email, role FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, ID)

	var user struct {
		ID    int
		Name  sql.NullString
		Email string
		Role  sql.NullString
	}
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by ID: %v\n", err)
		return nil, err
	}

	return &models.User{
		ID:    user.ID,
		Name:  user.Name.String,
		Email: user.Email,
		Role:  user.Role.String,
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
		user.Name = name.String
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
