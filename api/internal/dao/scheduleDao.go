package dao

import (
	"CapIot-api/internal/models"
	"context"
)

// ScheduleDAO defines the interface for scheduling data operations.
type ScheduleDAO interface {
	// Recurring Schedule Operations
	CreateRecurringSchedule(ctx context.Context, schedule models.RecurringSchedule) (int, error)
	GetRecurringScheduleByID(ctx context.Context, scheduleID int) (*models.RecurringSchedule, error)
	GetRecurringSchedulesByDevice(ctx context.Context, deviceID string) ([]models.RecurringSchedule, error)
	UpdateRecurringSchedule(ctx context.Context, scheduleID int, schedule models.RecurringSchedule) error
	DeleteRecurringSchedule(ctx context.Context, scheduleID int) error
}
