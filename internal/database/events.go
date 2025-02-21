package database

import (
	"calenduh-backend/internal/sqlc"
	"context"
	"fmt"
)

func CreateEvent(ctx context.Context, event sqlc.Event) error {
	query := `INSERT INTO events (event_id, calendar_id, name, start_time, end_time, location, description, notification, frequency, priority)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := DB.ExecContext(ctx, query, event.EventID, event.CalendarID, event.Name, 
		event.StartTime, event.EndTime, event.Location, event.Description, 
		event.Notification, event.Frequency, event.Priority)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

func GetEvent(ctx context.Context, eventID string) (*sqlc.Event, error) {
	query := `SELECT event_id, calendar_id, name, start_time, end_time, location, description, notification, frequency, priority 
			  FROM events WHERE event_id = $1`

	var event sqlc.Event
	err := DB.QueryRowContext(ctx, query, eventID).Scan(
		&event.EventID, &event.CalendarID, &event.Name, &event.StartTime, &event.EndTime, 
		&event.Location, &event.Description, &event.Notification, &event.Frequency, 
		&event.Priority,)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &event, nil
}

func UpdateEvent(ctx context.Context, event sqlc.Event) error {
	query := `UPDATE events SET calendar_id = $1, name = $2, start_time = $3, end_time = $4, location = $5, description = $6, notification = $7, frequency = $8, priority = $9 
			  WHERE event_id = $10`

	_, err := DB.ExecContext(ctx, query, event.CalendarID, event.Name, 
		event.StartTime, event.EndTime, event.Location, event.Description,
		event.Notification, event.Frequency, event.Priority, event.EventID)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func DeleteEvent(ctx context.Context, eventID string) error {
	query := `DELETE FROM events WHERE event_id = $1`
	_, err := DB.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}
