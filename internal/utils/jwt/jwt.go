package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/JSONStatham/sso/internal/domain/model"
	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user model.User, app model.App, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":    user.ID,
		"app_id": app.ID,
		"exp":    time.Now().Add(duration).Unix(),
	})

	fmt.Printf("LOC: %s", os.Getenv("JWT_SECRET"))

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
