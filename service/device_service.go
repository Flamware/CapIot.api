package service

import (
	"api.cap.iot/dao"
	"api.cap.iot/models"
)

type DefaultDeviceService struct {
	dao dao.DeviceDAO
}

type DeviceService interface {
	CreateDevice(device models.Device) error
	GetAllDevices() ([]models.Device, error)
}

func NewDeviceService(dao dao.DeviceDAO) *DefaultDeviceService {
	return &DefaultDeviceService{dao: dao}
}

func (s *DefaultDeviceService) CreateDevice(device models.Device) error {
	exists, err := s.dao.DeviceExists(device.DeviceID)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return s.dao.InsertDevice(device)
}

func (s *DefaultDeviceService) GetAllDevices() ([]models.Device, error) {
	return s.dao.GetAllDevices()
}
