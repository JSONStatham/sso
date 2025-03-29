package tests

import (
	"os"
	"testing"
	"time"

	ssov1 "github.com/JSONStatham/protos/gen/go/sso"
	"github.com/JSONStatham/sso/tests/suite"
	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	passDefaultLen = 10
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	// Test data
	email := gofakeit.Email()
	password := generatePassword()
	appID := st.GetTestAppID()

	// Registration test
	t.Run("Successful registration", func(t *testing.T) {
		registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
			Email:    email,
			Password: password,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, registerResponse.GetUserId())
	})

	// Login test
	t.Run("Successful login", func(t *testing.T) {
		loginTime := time.Now()

		loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
			Email:    email,
			Password: password,
			AppId:    int32(appID),
		})
		require.NoError(t, err)
		assert.NotEmpty(t, loginResponse.GetToken())

		// Verify JWT token
		verifyJWTToken(t, loginResponse.GetToken(), loginTime, st.Cfg.TokenTTL)
	})
}

func TestRegister_InvalidInput(t *testing.T) {
	ctx, st := suite.New(t)

	testCases := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Empty email",
			email:       "",
			password:    generatePassword(),
			expectedErr: "invalid email",
		},
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "invalid password",
		},
		{
			name:        "Invalid email format",
			email:       "not-an-email",
			password:    generatePassword(),
			expectedErr: "invalid email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tc.email,
				Password: tc.password,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func generatePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func verifyJWTToken(t *testing.T, token string, issuedAt time.Time, ttl time.Duration) {
	t.Helper()

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Verify standard claims
	require.NotEmpty(t, claims["exp"])
	require.NotEmpty(t, claims["iat"])

	// Verify custom claims
	require.NotEmpty(t, claims["uid"])
	require.NotEmpty(t, claims["app_id"])

	// Verify expiration time
	exp := time.Unix(int64(claims["exp"].(float64)), 0)
	assert.InDelta(t, issuedAt.Add(ttl).Unix(), exp.Unix(), 1)
}
