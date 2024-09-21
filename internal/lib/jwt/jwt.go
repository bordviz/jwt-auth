package jwt

import (
	"fmt"
	"jwt-auth/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func CreateToken(userID uuid.UUID, ipAddress string, refreshID int, secret string, expires time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["ip_address"] = ipAddress
	claims["refresh_id"] = refreshID
	claims["exp"] = time.Now().Add(expires).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DecodeToken(token string, secret string) (models.Token, error) {
	var model models.Token

	jwtToken, err := jwt.ParseWithClaims(token, &model, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !jwtToken.Valid {
		return models.Token{}, fmt.Errorf("failed to decode token: %w", err)
	}

	return model, nil
}
