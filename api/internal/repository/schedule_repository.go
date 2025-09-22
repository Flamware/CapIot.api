// File: repository/postgres_schedule.go

package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"log"
	"strings"
)

// PostgresScheduleDAO implements the ScheduleDAO interface using PostgreSQL.
type PostgresScheduleDAO struct {
	db *sql.DB
}

// NewPostgresScheduleDAO creates a new PostgresScheduleDAO.
func NewPostgresScheduleDAO(db *sql.DB) *PostgresScheduleDAO {
	return &PostgresScheduleDAO{db: db}
}

// CreateRecurringSchedule inserts a new recurring schedule into the database.
func (r *PostgresScheduleDAO) CreateRecurringSchedule(ctx context.Context, schedule models.RecurringSchedule) (int, error) {
	// 1. Détecter le type de récurrence et définir la priorité
	var priority int
	if strings.Contains(schedule.RecurrenceRule, "FREQ=MONTHLY") {
		priority = 100
	} else if strings.Contains(schedule.RecurrenceRule, "FREQ=WEEKLY") {
		priority = 10
	} else if strings.Contains(schedule.RecurrenceRule, "FREQ=DAILY") {
		priority = 1
	} else if strings.Contains(schedule.RecurrenceRule, "FREQ=ONCE") {
		priority = 1000
	} else {
		priority = 0 // Valeur par défaut si le type de récurrence n'est pas reconnu
	}
	// 2. Mettre à jour la requête SQL pour inclure la colonne priority
	query := `
        INSERT INTO public.recurring_schedules (
            device_id, schedule_name, start_time, end_time, recurrence_rule, priority, is_exception
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING recurring_schedule_id
    `
	var scheduleID int
	err := r.db.QueryRowContext(
		ctx,
		query,
		schedule.DeviceID,
		schedule.ScheduleName,
		schedule.StartTime,
		schedule.EndTime,
		schedule.RecurrenceRule,
		priority, // 3. Ajouter la valeur de priorité ici
		schedule.IsException,
	).Scan(&scheduleID)

	if err != nil {
		log.Printf("Error creating recurring schedule: %v", err)
		return 0, err
	}

	return scheduleID, nil
}

// GetRecurringScheduleByID retrieves a recurring schedule by its ID.
func (r *PostgresScheduleDAO) GetRecurringScheduleByID(ctx context.Context, scheduleID int) (*models.RecurringSchedule, error) {
	query := `
        SELECT recurring_schedule_id, device_id, schedule_name, start_time, end_time, recurrence_rule, is_exception
        FROM public.recurring_schedules
        WHERE recurring_schedule_id = $1
    `
	var s models.RecurringSchedule
	err := r.db.QueryRowContext(ctx, query, scheduleID).Scan(
		&s.RecurringScheduleID,
		&s.DeviceID,
		&s.ScheduleName,
		&s.StartTime,
		&s.EndTime,
		&s.RecurrenceRule,
		&s.IsException,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error getting recurring schedule by ID: %v", err)
		return nil, err
	}
	return &s, nil
}

// GetRecurringSchedulesByDevice retrieves all recurring schedules for a given device.
func (r *PostgresScheduleDAO) GetRecurringSchedulesByDevice(ctx context.Context, deviceID string) ([]models.RecurringSchedule, error) {
	query := `
        SELECT recurring_schedule_id, device_id, schedule_name, start_time, end_time, recurrence_rule, is_exception
		FROM public.recurring_schedules
		WHERE device_id = $1
		ORDER BY priority DESC, start_time ASC
				`
	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		log.Printf("Error getting recurring schedules by device: %v", err)
		return nil, err
	}
	defer rows.Close()

	var schedules []models.RecurringSchedule
	for rows.Next() {
		var s models.RecurringSchedule
		if err := rows.Scan(
			&s.RecurringScheduleID,
			&s.DeviceID,
			&s.ScheduleName,
			&s.StartTime,
			&s.EndTime,
			&s.RecurrenceRule,
			&s.IsException,
		); err != nil {
			log.Printf("Error scanning recurring schedule row: %v", err)
			return nil, err
		}
		schedules = append(schedules, s)
	}
	log.Printf("Retrieved %d schedules for device %s", len(schedules), deviceID)
	log.Printf("Schedules: %+v", schedules)
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating recurring schedule rows: %v", err)
		return nil, err
	}

	return schedules, nil
}

// UpdateRecurringSchedule updates an existing recurring schedule.
func (r *PostgresScheduleDAO) UpdateRecurringSchedule(ctx context.Context, scheduleID int, schedule models.RecurringSchedule) error {
	query := `
        UPDATE public.recurring_schedules
        SET device_id = $2, schedule_name = $3, start_time = $4, end_time = $5, recurrence_rule = $6
        WHERE recurring_schedule_id = $1
    `
	result, err := r.db.ExecContext(
		ctx,
		query,
		scheduleID,
		schedule.DeviceID,
		schedule.ScheduleName,
		schedule.StartTime,
		schedule.EndTime,
		schedule.RecurrenceRule,
	)
	if err != nil {
		log.Printf("Error updating recurring schedule: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteRecurringSchedule deletes a recurring schedule by its ID.
func (r *PostgresScheduleDAO) DeleteRecurringSchedule(ctx context.Context, scheduleID int) error {
	query := `
        DELETE FROM public.recurring_schedules
        WHERE recurring_schedule_id = $1
    `
	result, err := r.db.ExecContext(ctx, query, scheduleID)
	if err != nil {
		log.Printf("Error deleting recurring schedule: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
