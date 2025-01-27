package models

import (
	"time"

	"github.com/iamnande/hyrule/internal/repositories/user"
	"github.com/segmentio/ksuid"
)

type User struct {
	ID ksuid.KSUID

	Email    string
	FullName string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func MarshalUser(record *user.Record) User {
	return User{
		ID:        record.ID,
		Email:     record.Email,
		FullName:  record.FullName,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		DeletedAt: record.DeletedAt,
	}
}
