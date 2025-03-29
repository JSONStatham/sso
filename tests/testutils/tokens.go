package testutils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateExpiredToken(userID int64, appID int32) (string, error) {
	// Create token with expiration set to 1 hour ago
	claims := jwt.MapClaims{
		"uid":    userID,
		"app_id": appID,
		"exp":    time.Now().Add(-1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
