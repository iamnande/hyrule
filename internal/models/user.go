package models

import (
	"time"

	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/repositories/users"
)

type User struct {
	ID ksuid.KSUID

	Email    string
	FullName string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func MarshalUser(record *users.Record) User {
	return User{
		ID:        record.ID,
		Email:     record.Email,
		FullName:  record.FullName,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}
