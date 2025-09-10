package dao

import (
	"CapIot-api/internal/models"
	"context"
)

type NotificationDAO interface {
	GetNotifications(ctx context.Context, userID int, limit int, offset int) ([]models.Notification, error)
	CountNotifications(ctx context.Context, userID int) (int, error)
	MarkAllAsRead(ctx context.Context, userID int) error
	MarkAsRead(ctx context.Context, notificationID int, userID int) error
	DeleteNotification(ctx context.Context, notificationID int, userID int) error
	CreateNotification(ctx context.Context, notification models.Notification) (int, error)
	DeleteAllNotifications(ctx context.Context, id int) error
}
