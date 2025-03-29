package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/JSONStatham/sso/internal/config"
	"github.com/JSONStatham/sso/internal/domain/model"
	"github.com/JSONStatham/sso/internal/storage"
	"github.com/JSONStatham/sso/internal/utils/jwt"
	"github.com/JSONStatham/sso/internal/utils/logger/sl"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid creadentials")
	ErrInvalidAppID       = errors.New("invalid application id")
)

type Auth struct {
	log *slog.Logger
	cfg *config.Config
	st  Storage
}

type Storage interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	User(ctx context.Context, email string) (model.User, error)
	IsAdmin(ctx context.Context, uid int64) (bool, error)
	App(ctx context.Context, appID int64) (model.App, error)
}

func New(log *slog.Logger, cfg *config.Config, st Storage) *Auth {
	return &Auth{
		log: log,
		cfg: cfg,
		st:  st,
	}
}

func (a *Auth) RegisterUser(ctx context.Context, email, password string) (int64, error) {
	const op = "auth.RegisterUser"

	log := a.log.With(slog.String("op", op))

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	uid, err := a.st.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered", slog.Int64("uid", uid), slog.String("email", email))

	return uid, nil
}

func (a *Auth) Login(ctx context.Context, email, password string, appID int64) (string, error) {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op))

	log.Info("attempting to login user")

	user, err := a.st.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		log.Warn("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.st.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		log.Error("failed to get app", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in succefffully", slog.Int("uid", user.ID), slog.String("email", email))

	token, err := jwt.NewToken(user, app, a.cfg.TokenTTL)
	if err != nil {
		log.Error("failed to create jwt token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) Logout(ctx context.Context, token string) error {
	panic("not implemented")
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(slog.String("op", op), slog.Int64("uid", userID))

	isAdmin, err := a.st.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
