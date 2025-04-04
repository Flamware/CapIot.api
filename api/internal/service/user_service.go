package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
)

type DefaultUserService struct {
	dao dao.UserDAO
}

type UserService interface {
	CreateUser(user models.User) error
	GetUserByID(id int) (*models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id int) error
	GetAllUsers() ([]models.User, error) // Add this method to the interface
}

func NewUserService(dao dao.UserDAO) *DefaultUserService {
	return &DefaultUserService{dao: dao}
}

func (s *DefaultUserService) CreateUser(auth0_id string, auth0_email string) (int, error) {
	userID, err := s.dao.UserExists(auth0_id)
	if err != nil {
		return 0, err
	}

	if userID != 0 {
		return userID, nil
	}

	userID, err = s.dao.CreateUser(auth0_id, auth0_email)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *DefaultUserService) GetUserByID(id int) (*models.User, error) {
	return s.dao.FindUserByID(id)
}

func (s *DefaultUserService) UpdateUser(user models.User) (*models.User, error) {
	return s.dao.UpdateUser(user)
}

func (s *DefaultUserService) DeleteUser(id int) error {
	return s.dao.DeleteUser(id)
}

func (s *DefaultUserService) GetAllUsers() ([]models.User, error) { // Correct the return type
	return s.dao.GetAllUsers()
}
