package jwt

import (
	"os"
	"testing"
	"time"

	"github.com/JSONStatham/sso/internal/domain/model"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	defer os.Unsetenv("JWT_SECRET")

	user := model.User{ID: 1}
	app := model.App{ID: 1}
	duration := time.Minute * 15

	tokenStr, err := NewToken(user, app, duration)
	require.NoError(t, err, "Toeken generation shopuld not return an error")
	require.NotEmpty(t, tokenStr, "Token string should not be empty")

	// Parse token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err, "Token parsing should not return an error")
	require.NotNil(t, token, "Parsed token should not be nil")

	// Validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok, "Token claims should be of type jwt.MapClaims")

	assert.Equal(t, float64(user.ID), claims["uid"], "Token claims should contain user ID")
	assert.Equal(t, float64(app.ID), claims["app_id"], "Token claims should contain app ID")

	exp, ok := claims["exp"].(float64)
	require.True(t, ok, "Exp time should be a float64")
	assert.Greater(t, exp, float64(time.Now().Unix()), "Exp time should be in the future")
}
