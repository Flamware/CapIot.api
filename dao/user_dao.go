package dao

import (
	"api.cap.iot/models"
	"database/sql"
	"errors"
	"log"
)

// UserDAO struct définit les méthodes d'accès aux données des utilisateurs
type UserDAO struct {
	DB *sql.DB
}

// NewUserDAO retourne une nouvelle instance de UserDAO avec une connexion DB
func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{DB: db}
}

// CreateUser insère un utilisateur dans la base de données PostgreSQL
func (dao *UserDAO) CreateUser(user *models.User) (string, error) {
	sqlStatement := `
		INSERT INTO users (email, password, role) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`
	var email string
	err := dao.DB.QueryRow(sqlStatement, user.Email, user.Password, user.Role).Scan(&email)
	if err != nil {
		log.Printf("❌ Error inserting user: %v", err)
		return "", err
	}

	return email, nil
}

// GetUserByEmail retrieves a user by their email
func (dao *UserDAO) GetUserByEmail(email string) (*models.User, error) {
	sqlStatement := `SELECT id, email, name, password, role, created_at FROM users WHERE email=$1`
	var user models.User
	log.Printf("Executing query to fetch user with email: %s", email) // Log email being queried

	// Use pointers for nullable fields like `name`
	var name *string // Use a pointer to handle NULL values in the 'name' column
	err := dao.DB.QueryRow(sqlStatement, email).Scan(&user.ID, &user.Email, &name, &user.Password, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("No user found with the provided email.")
			return nil, nil // Return nil if no user is found
		}
		log.Printf("Error fetching user by email: %v", err) // Log any other errors
		return nil, err
	}

	// If name is not NULL, assign the value
	if name != nil {
		user.Name = *name
	} else {
		user.Name = "" // Set an empty string if the name is NULL
	}

	log.Printf("User found: %v", user) // Log the user details
	return &user, nil
}
