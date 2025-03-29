package tests

import (
	"context"
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
	emptyAppID     = 0
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email, password := registerNewUser(ctx, t, st.AuthClient)

	// Login test
	loginTime := time.Now()

	loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    st.GetTestAppID(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginResponse.GetToken())

	// Verify JWT token
	claims := verifyJWTToken(t, loginResponse.GetToken())

	// Verify expiration time
	exp := time.Unix(int64(claims["exp"].(float64)), 0)
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), exp.Unix(), 1)
}

func Test__CannotRegisterTwice(t *testing.T) {
	ctx, st := suite.New(t)
	email, password := registerNewUser(ctx, t, st.AuthClient)

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, registerResponse.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func Test_RegisterInvalidInput(t *testing.T) {
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
			expectedErr: "Field validation for 'Email' failed on the 'required'",
		},
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "Field validation for 'Password' failed on the 'required' ",
		},
		{
			name:        "Invalid email format",
			email:       "not-an-email",
			password:    generatePassword(),
			expectedErr: "Field validation for 'Email' failed on the 'email' ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tc.email,
				Password: tc.password,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expectedErr)
		})
	}
}

func Test_LoginInvalidInput(t *testing.T) {
	ctx, st := suite.New(t)

	testCases := []struct {
		name        string
		email       string
		password    string
		appID       int64
		expectedErr string
	}{
		{
			name:        "Empty email",
			email:       "",
			password:    generatePassword(),
			appID:       st.GetTestAppID(),
			expectedErr: "Field validation for 'Email' failed on the 'required' tag",
		},
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       st.GetTestAppID(),
			expectedErr: "Field validation for 'Password' failed on the 'required' tag",
		},
		{
			name:        "Invalid email format",
			email:       "not-an-email",
			password:    generatePassword(),
			appID:       st.GetTestAppID(),
			expectedErr: "Field validation for 'Email' failed on the 'email' tag",
		},
		{
			name:        "Empty input",
			email:       "",
			password:    "",
			appID:       st.GetTestAppID(),
			expectedErr: "'LoginRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag\nKey: 'LoginRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag",
		},
		{
			name:        "Invalid App ID",
			email:       gofakeit.Email(),
			password:    generatePassword(),
			appID:       emptyAppID,
			expectedErr: "Error:Field validation for 'AppID' failed on the 'required' tag",
		},
		{
			name:        "User not found",
			email:       gofakeit.Email(),
			password:    generatePassword(),
			appID:       st.GetTestAppID(),
			expectedErr: "user not found",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: generatePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func registerNewUser(ctx context.Context, t *testing.T, client ssov1.AuthClient) (email, password string) {
	email = gofakeit.Email()
	password = generatePassword()

	registerResponse, err := client.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, registerResponse.GetUserId())
	return email, password
}

func generatePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func verifyJWTToken(t *testing.T, token string) jwt.MapClaims {
	t.Helper()

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Verify standard claims
	require.NotEmpty(t, claims["exp"])

	// Verify custom claims
	require.NotEmpty(t, claims["uid"])
	require.NotEmpty(t, claims["app_id"])

	return claims
}
