// File: internal/models/schedule.go

package models

import (
	"time"
)

// RecurringSchedule represents a recurring schedule from the 'recurring_schedules' table.
type RecurringSchedule struct {
	RecurringScheduleID int       `json:"recurring_schedule_id"`
	DeviceID            string    `json:"device_id"`
	ScheduleName        string    `json:"schedule_name"`
	StartTime           time.Time `json:"start_time"` // This contains both the date and the time
	EndTime             time.Time `json:"end_time"`   // This contains both the date and the time
	RecurrenceRule      string    `json:"recurrence_rule"`
	IsException         bool      `json:"is_exception"`
}
