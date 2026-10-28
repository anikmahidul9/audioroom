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
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Users:  &UserStore{db},
		Events: &EventStore{db},
	}
}
