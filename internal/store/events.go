package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrInvalidEventTime = errors.New("scheduled_at must be in the future")
)

type Event struct {
	ID          uuid.UUID  `json:"id"`
	HostID      uuid.UUID  `json:"host_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CategoryID  *int       `json:"category_id,omitempty"`
	Visibility  string     `json:"visibility"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	RoomName    string     `json:"room_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type EventStore struct {
	db *sql.DB
}

func (s *EventStore) Create(ctx context.Context, e *Event) (*Event, error) {
	const query = `
		INSERT INTO events
			(host_id, title, description, category_id, visibility, status, scheduled_at, room_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, host_id, title, description, category_id, visibility, status,
		          scheduled_at, started_at, ended_at, room_name, created_at, updated_at
	`
	row := s.db.QueryRowContext(ctx, query,
		e.HostID, e.Title, e.Description, e.CategoryID,
		e.Visibility, e.Status, e.ScheduledAt, e.RoomName,
	)
	return scanEvent(row)
}

func (s *EventStore) GetByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	const query = `
		SELECT id, host_id, title, description, category_id, visibility, status,
		       scheduled_at, started_at, ended_at, room_name, created_at, updated_at
		FROM events WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, id)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEventNotFound
	}
	return e, err
}

func (s *EventStore) ListPublic(ctx context.Context, status string, limit, offset int) ([]*Event, error) {
	const query = `
		SELECT id, host_id, title, description, category_id, visibility, status,
		       scheduled_at, started_at, ended_at, room_name, created_at, updated_at
		FROM events
		WHERE visibility = 'public' AND status = $1
		ORDER BY
		  CASE WHEN $1 = 'scheduled' THEN scheduled_at END ASC NULLS LAST,
		  CASE WHEN $1 = 'live' THEN started_at END DESC NULLS LAST
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *EventStore) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	var query string
	switch status {
	case "live":
		query = `UPDATE events SET status = 'live', started_at = NOW() WHERE id = $1`
	case "ended":
		query = `UPDATE events SET status = 'ended', ended_at = NOW() WHERE id = $1`
	default:
		query = `UPDATE events SET status = $2 WHERE id = $1`
	}
	res, err := s.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrEventNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(sc scanner) (*Event, error) {
	e := &Event{}
	err := sc.Scan(
		&e.ID, &e.HostID, &e.Title, &e.Description, &e.CategoryID,
		&e.Visibility, &e.Status, &e.ScheduledAt, &e.StartedAt, &e.EndedAt,
		&e.RoomName, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return e, nil
}