package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type EventMember struct {
	ID       uuid.UUID `json:"id"`
	EventID  uuid.UUID `json:"event_id"`
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	Status   string    `json:"status"`
	JoinedAt time.Time `json:"joined_at"`
}

type EventMemberStore struct {
	db *sql.DB
}

func (s *EventMemberStore) Join(ctx context.Context, eventID, userID uuid.UUID, role string) error {
	const query = `
		INSERT INTO event_members (event_id, user_id, role, status)
		VALUES ($1, $2, $3, 'active')
		ON CONFLICT (event_id, user_id) 
		DO UPDATE SET status = 'active', role = EXCLUDED.role
	`
	_, err := s.db.ExecContext(ctx, query, eventID, userID, role)
	return err
}

func (s *EventMemberStore) Leave(ctx context.Context, eventID, userID uuid.UUID) error {
	const query = `UPDATE event_members SET status = 'left' WHERE event_id = $1 AND user_id = $2`
	_, err := s.db.ExecContext(ctx, query, eventID, userID)
	return err
}

func (s *EventMemberStore) UpdateRole(ctx context.Context, eventID, userID uuid.UUID, newRole string) error {
	const query = `UPDATE event_members SET role = $3 WHERE event_id = $1 AND user_id = $2`
	_, err := s.db.ExecContext(ctx, query, eventID, userID, newRole)
	return err
}

func (s *EventMemberStore) ListActive(ctx context.Context, eventID uuid.UUID) ([]*EventMember, error) {
	const query = `
		SELECT m.id, m.event_id, m.user_id, u.name, m.role, m.status, m.joined_at
		FROM event_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.event_id = $1 AND m.status = 'active'
		ORDER BY 
			CASE m.role 
				WHEN 'host' THEN 1 
				WHEN 'co_host' THEN 2 
				WHEN 'speaker' THEN 3 
				ELSE 4 
			END
	`
	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*EventMember, 0)
	for rows.Next() {
		m := &EventMember{}
		err := rows.Scan(&m.ID, &m.EventID, &m.UserID, &m.Name, &m.Role, &m.Status, &m.JoinedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}
