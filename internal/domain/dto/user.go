package dto

import (
	"fmt"
	"jwt-auth/internal/lib/validator"
	"strings"
)

type User struct {
	Email string `json:"email" validate:"required,email"`
}

func (u *User) Validate() error {
	u.Email = strings.TrimSpace(u.Email)

	if err := validator.Validate(u); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}

	return nil
}
