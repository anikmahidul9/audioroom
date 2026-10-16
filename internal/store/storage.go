package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Users interface {
		Create(ctx context.Context, user *User) error
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Users: &UserStore{db},
	}
}
