package repository

import (
	"api.cap.iot/dao"
	"api.cap.iot/models"
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

// CreateUser inserts a user with the Auth0 ID and email into the database.
func (r *PostgresUserRepository) CreateUser(auth0ID string, auth0_email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO users (auth0_id, email) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, auth0ID, auth0_email)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		return err
	}
	log.Printf("User with Auth0 ID %s inserted into database.\n", auth0ID)
	return nil
}

// UserExists checks if a user exists in the database.
func (r *PostgresUserRepository) UserExists(auth0_id string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const query = `SELECT 1 FROM users WHERE auth0_id = $1 LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(ctx, query, auth0_id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Printf("Error checking user existence: %v\n", err)
		return false, err
	}
	return true, nil
}

// FindUserByID retrieves a user by their ID from the database.
func (r *PostgresUserRepository) FindUserByID(ID int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, ID)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error fetching user by ID: %v\n", err)
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates a user in the database.
func (r *PostgresUserRepository) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET name = $1, email = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		log.Printf("Error updating user in database: %v\n", err)
		return err
	}
	log.Printf("User %s updated in database.\n", user.ID)
	return nil
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
