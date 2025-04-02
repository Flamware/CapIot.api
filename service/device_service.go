package service

import (
	"api.cap.iot/dao"
	"api.cap.iot/models"
	errors "errors"
	"log"
)

type DefaultDeviceService struct {
	dao dao.DeviceDAO
}

type DeviceService interface {
	CreateDevice(device models.Device) error
	GetAllDevices() ([]models.Device, error)
	SetDeviceToLocation(device_id string, location_id int) interface{}
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
		log.Printf("Device already exists: %s", device.DeviceID)
		return nil
	}

	log.Printf("Creating device: %s", device.DeviceID)
	return s.dao.InsertDevice(device)
}

func (s *DefaultDeviceService) GetAllDevices() ([]models.Device, error) {
	return s.dao.GetAllDevices()
}

func (s *DefaultDeviceService) SetDeviceToLocation(device_id string, location_id int) error {
	assigned, err := s.dao.IsDeviceAssigned(device_id)
	if err != nil {
		return err
	}

	if assigned {
		log.Printf("Device %s is already assigned to a location", device_id)
		return errors.New("device is already assigned to a location")
	}

	return s.dao.SetDeviceToLocation(device_id, location_id)
}
