package models

import (
	"github.com/iamnande/hyrule/internal/repositories/user"
	"github.com/iamnande/hyrule/internal/rest/apis/response"
)

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func MarshalUser(record *user.Record) *User {
	return &User{
		ID:        response.UUID(record.ID),
		FirstName: record.FirstName,
		LastName:  record.LastName,
		Email:     response.Email(record.Email),
		CreatedAt: response.Time(record.CreatedAt),
		UpdatedAt: response.Time(record.UpdatedAt),
	}
}
