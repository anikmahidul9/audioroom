package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, user *User) error
	}
	Events interface {
		Create(ctx context.Context, e *Event) (*Event, error)
		GetByID(ctx context.Context, id uuid.UUID) (*Event, error)
		ListPublic(ctx context.Context, status string, limit, offset int) ([]*Event, error)
		UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	}
	Members interface {
		Join(ctx context.Context, eventID, userID uuid.UUID, role string) error
		Leave(ctx context.Context, eventID, userID uuid.UUID) error
		UpdateRole(ctx context.Context, eventID, userID uuid.UUID, newRole string) error
		ListActive(ctx context.Context, eventID uuid.UUID) ([]*EventMember, error)
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Users:  &UserStore{db},
		Events: &EventStore{db},
		Members: &EventMemberStore{db},
	}
}
