package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Token struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID `json:"user_id"`
	IpAddress string    `json:"ip_address"`
	RefreshID int       `json:"refresh_id"`
}

type TokensPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
