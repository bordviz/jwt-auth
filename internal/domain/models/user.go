package models

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type UserWithRefresh struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	RefreshID int       `json:"refresh_id"`
}
