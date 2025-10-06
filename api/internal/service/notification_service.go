package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"fmt"
)

type DefaultNotificationService struct {
	NotificationDAO dao.NotificationDAO
}

// NewNotificationService creates a new DefaultNotificationService instance
func NewNotificationService(NotificationDAO dao.NotificationDAO) *DefaultNotificationService {
	return &DefaultNotificationService{
		NotificationDAO: NotificationDAO,
	}
}

// NotificationService defines the interface for Notification-related business logic
type NotificationService interface {
	GetNotifications(ctx context.Context, userID int, page int, limit int) (map[string]interface{}, error)
	CountNotifications(ctx context.Context, userID int) (int, error)
	MarkAllAsRead(ctx context.Context, userID int) error
	MarkAsRead(ctx context.Context, notificationID int, userID int) error
	DeleteNotification(ctx context.Context, notificationID int, userID int) error
	CreateNotification(ctx context.Context, notification models.Notification) (int, error)
	DeleteAllNotifications(ctx context.Context, id int) error
	GetDeviceNotifications(ctx context.Context, id string, page int, limit int) (map[string]interface{}, error) // Changed 'string' to 'int'
}

// GetNotifications retrieves notifications for a user with pagination
func (s *DefaultNotificationService) GetNotifications(ctx context.Context, userID, page, limit int) (map[string]interface{}, error) {
	notifications, err := s.NotificationDAO.GetNotifications(ctx, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications for user with ID %d: %w", userID, err)
	}

	totalNotifications, err := s.NotificationDAO.CountNotifications(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications for user with ID %d: %w", userID, err)
	}

	totalPages := (totalNotifications + limit - 1) / limit
	response := map[string]interface{}{
		"data":        notifications,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalNotifications,
		"totalPages":  totalPages,
	}
	return response, nil
}

// CountNotifications counts the total number of notifications for a user
func (s *DefaultNotificationService) CountNotifications(ctx context.Context, userID int) (int, error) {
	return s.NotificationDAO.CountNotifications(ctx, userID)
}

// MarkAllAsRead marks all notifications as read for a given user
func (s *DefaultNotificationService) MarkAllAsRead(ctx context.Context, userID int) error {
	if err := s.NotificationDAO.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read for user %d: %w", userID, err)
	}
	return nil
}

// MarkAsRead marks a specific notification as read for a given user
func (s *DefaultNotificationService) MarkAsRead(ctx context.Context, notificationID int, userID int) error {
	if err := s.NotificationDAO.MarkAsRead(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("failed to mark notification %d as read for user %d: %w", notificationID, userID, err)
	}
	return nil
}

// DeleteNotification deletes a specific notification for a given user
func (s *DefaultNotificationService) DeleteNotification(ctx context.Context, notificationID int, userID int) error {
	if err := s.NotificationDAO.DeleteNotification(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("failed to delete notification %d for user %d: %w", notificationID, userID, err)
	}
	return nil
}

// CreateNotification creates a new notification
func (s *DefaultNotificationService) CreateNotification(ctx context.Context, notification models.Notification) (int, error) {
	id, err := s.NotificationDAO.CreateNotification(ctx, notification)
	if err != nil {
		return 0, fmt.Errorf("failed to create notification: %w", err)
	}
	return id, nil
}

// DeleteAllNotifications deletes all notifications for a given user
func (s *DefaultNotificationService) DeleteAllNotifications(ctx context.Context, userID int) error {
	if err := s.NotificationDAO.DeleteAllNotifications(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete all notifications for user %d: %w", userID, err)
	}
	return nil
}

// GetDeviceNotifications retrieves notifications for a specific device with pagination
func (s *DefaultNotificationService) GetDeviceNotifications(ctx context.Context, deviceID string, page int, limit int) (map[string]interface{}, error) {
	notifications, err := s.NotificationDAO.GetDeviceNotifications(ctx, deviceID, limit, (page-1)*limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications for device with ID %d: %w", deviceID, err)
	}

	totalNotifications, err := s.NotificationDAO.CountDeviceNotifications(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications for device with ID %d: %w", deviceID, err)
	}

	totalPages := (totalNotifications + limit - 1) / limit
	response := map[string]interface{}{
		"data":        notifications,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalNotifications,
		"totalPages":  totalPages,
	}
	return response, nil
}
