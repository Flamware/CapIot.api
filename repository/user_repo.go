package repository

import (
	"api.cap.iot/dao"
	"api.cap.iot/models"
	"fmt"
)

// UserRepository gère les opérations métiers liées aux utilisateurs
type UserRepository struct {
	UserDAO *dao.UserDAO
}

// NewUserRepository crée une nouvelle instance de UserRepository
func NewUserRepository(userDAO *dao.UserDAO) *UserRepository {
	return &UserRepository{
		UserDAO: userDAO,
	}
}

// CreateUser appelle le DAO pour insérer un utilisateur et renvoie son ID
func (repo *UserRepository) CreateUser(user *models.User) (string, error) {
	email, err := repo.UserDAO.CreateUser(user)
	if err != nil {
		return "", fmt.Errorf("❌ Error creating user: %v", err)
	}
	return email, nil
}

// GetUserByEmail appelle le DAO pour récupérer un utilisateur par email
func (repo *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	user, err := repo.UserDAO.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("❌ Error fetching user by email: %v", err)
	}
	return user, nil
}
