package service

import (
	"api.cap.iot/models"
	repositories "api.cap.iot/repository"
	"log"
)

type DeviceService interface {
	CreateDevice(device models.Device) error
}

type DefaultDeviceService struct {
	repo repositories.DeviceRepository
}

func NewDeviceService(repo repositories.DeviceRepository) *DefaultDeviceService {
	return &DefaultDeviceService{repo: repo}
}

func (s *DefaultDeviceService) CreateDevice(device models.Device) error {
	exists, err := s.repo.DeviceExists(device.DeviceID)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("Device %s already exists in the database.\n", device.DeviceID)
		return nil // Or return an error if you want to indicate a duplicate
	}

	return s.repo.InsertDevice(device)
}
