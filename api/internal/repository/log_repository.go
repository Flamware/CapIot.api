package repository

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
)

// PostgresNotificationDAO implements the NotificationDAO interface using PostgreSQL.
type PostgresNotificationDAO struct {
	db *sql.DB
}

// NewPostgresNotificationDAO creates a new PostgresNotificationDAO.
func NewPostgresNotificationDAO(db *sql.DB) *PostgresNotificationDAO {
	return &PostgresNotificationDAO{db: db}
}

// GetNotifications retrieves notifications for a user with pagination.
func (r *PostgresNotificationDAO) GetNotifications(ctx context.Context, userID, limit, offset int) ([]models.Notification, error) {
	query := `
       SELECT
           l.log_id,
           l.log_content,
           l.log_timestamp,
           l.log_read,
           s.site_id,
           s.site_name,
           loc.location_id,
           loc.location_name,
           c.component_name,
           c.component_type
       FROM
           public.component_log AS l
       JOIN public.components AS c ON l.component_id = c.component_id
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE
           us.user_id = $1
           AND dl.is_current = true
       ORDER BY l.log_timestamp DESC
       LIMIT $2 OFFSET $3
    `

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		log.Printf("Error querying notifications: %v", err)
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var n models.Notification

		if err := rows.Scan(
			&n.LogID,
			&n.LogContent,
			&n.LogTimestamp,
			&n.LogRead,
			&n.SiteID,
			&n.SiteName,
			&n.LocationID,
			&n.LocationName,
			&n.Component,
			&n.ComponentType,
		); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return notifications, nil
}

// CountNotifications returns the total number of notifications for a user.
func (r *PostgresNotificationDAO) CountNotifications(ctx context.Context, userID int) (int, error) {
	query := `
       SELECT COUNT(*)
       FROM public.component_log AS l
       JOIN public.components AS c ON l.component_id = c.component_id
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE us.user_id = $1 AND dl.is_current = true
    `

	var count int
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&count); err != nil {
		log.Printf("Error counting notifications: %v", err)
		return 0, err
	}

	return count, nil
}

// MarkAllAsRead sets log_read = true for all notifications of a user
func (r *PostgresNotificationDAO) MarkAllAsRead(ctx context.Context, userID int) error {
	query := `
       UPDATE public.component_log AS l
       SET log_read = true
       FROM public.components AS c
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE l.component_id = c.component_id
       AND us.user_id = $1
       AND dl.is_current = true
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// MarkAsRead sets log_read = true for a specific notification of a user
func (r *PostgresNotificationDAO) MarkAsRead(ctx context.Context, notificationID, userID int) error {
	query := `
       UPDATE public.component_log AS l
       SET log_read = true
       FROM public.components AS c
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE l.component_id = c.component_id
       AND l.log_id = $1
       AND us.user_id = $2
       AND dl.is_current = true
    `
	_, err := r.db.ExecContext(ctx, query, notificationID, userID)
	return err
}

// DeleteNotification removes a specific notification of a user
func (r *PostgresNotificationDAO) DeleteNotification(ctx context.Context, notificationID, userID int) error {
	query := `
       DELETE FROM public.component_log l
       USING public.components AS c
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE l.component_id = c.component_id
       AND l.log_id = $1
       AND us.user_id = $2
       AND dl.is_current = true
    `
	_, err := r.db.ExecContext(ctx, query, notificationID, userID)
	return err
}

// CreateNotification inserts a new notification and returns its ID
func (r *PostgresNotificationDAO) CreateNotification(ctx context.Context, notification models.Notification) (int, error) {
	query := `
       INSERT INTO public.component_log (component_id, log_content, log_timestamp, log_read)
       VALUES ($1, $2, $3, $4)
       RETURNING log_id
    `
	var logID int
	err := r.db.QueryRowContext(ctx, query,
		notification.Component,
		notification.LogContent,
		notification.LogTimestamp,
		notification.LogRead,
	).Scan(&logID)
	if err != nil {
		return 0, err
	}
	return logID, nil
}

// DeleteAllNotifications removes all notifications of a user
func (r *PostgresNotificationDAO) DeleteAllNotifications(ctx context.Context, userID int) error {
	query := `
       DELETE FROM public.component_log l
       USING public.components AS c
       JOIN public.devices AS d ON c.device_id = d.device_id
       JOIN public.device_location AS dl ON d.device_id = dl.device_id
       JOIN public.locations AS loc ON dl.location_id = loc.location_id
       JOIN public.sites AS s ON loc.site_id = s.site_id
       JOIN public.user_site AS us ON s.site_id = us.site_id
       WHERE l.component_id = c.component_id
       AND us.user_id = $1
       AND dl.is_current = true
    `
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// GetDeviceNotifications retrieves notifications for a specific device with pagination
func (r *PostgresNotificationDAO) GetDeviceNotifications(ctx context.Context, deviceID string, limit, offset int) ([]models.Notification, error) {
	// Check if deviceID exists
	queryCheck := `SELECT 1 FROM public.devices WHERE device_id = $1`
	var exists int
	err := r.db.QueryRowContext(ctx, queryCheck, deviceID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return a clear "not found" error
			return nil, fmt.Errorf("device with ID %s does not exist", deviceID)
		}
		log.Printf("Error checking device existence: %v", err)
		return nil, err
	}

	query := `
		SELECT
			l.log_id,
			l.log_content,
			l.log_timestamp,
			l.log_read,
			s.site_id,
			s.site_name,
			loc.location_id,
			loc.location_name,
			c.component_name,
			c.component_type
		FROM
			public.component_log AS l
		JOIN public.components AS c ON l.component_id = c.component_id
		JOIN public.devices AS d ON c.device_id = d.device_id
		JOIN public.device_location AS dl ON d.device_id = dl.device_id
		JOIN public.locations AS loc ON dl.location_id = loc.location_id
		JOIN public.sites AS s ON loc.site_id = s.site_id
		WHERE
			d.device_id = $1
			AND dl.is_current = true
		ORDER BY l.log_timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID, limit, offset)
	if err != nil {
		log.Printf("Error querying device notifications: %v", err)
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var n models.Notification

		if err := rows.Scan(
			&n.LogID,
			&n.LogContent,
			&n.LogTimestamp,
			&n.LogRead,
			&n.SiteID,
			&n.SiteName,
			&n.LocationID,
			&n.LocationName,
			&n.Component,
			&n.ComponentType,
		); err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return nil, err
	}

	return notifications, nil
}

// CountDeviceNotifications returns the total number of notifications for a specific device
func (r *PostgresNotificationDAO) CountDeviceNotifications(ctx context.Context, deviceID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM public.component_log AS l
		JOIN public.components AS c ON l.component_id = c.component_id
		JOIN public.devices AS d ON c.device_id = d.device_id
		JOIN public.device_location AS dl ON d.device_id = dl.device_id
		JOIN public.locations AS loc ON dl.location_id = loc.location_id
		JOIN public.sites AS s ON loc.site_id = s.site_id
		WHERE d.device_id = $1 AND dl.is_current = true
	`

	var count int
	if err := r.db.QueryRowContext(ctx, query, deviceID).Scan(&count); err != nil {
		log.Printf("Error counting device notifications: %v", err)
		return 0, err
	}

	return count, nil
}
