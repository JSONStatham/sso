package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JSONStatham/sso/internal/domain/model"
	"github.com/JSONStatham/sso/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePah string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePah)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	res, err := s.db.ExecContext(ctx, "INSERT INTO users (email, password) VALUES (?, ?)", email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	uid, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *Storage) User(ctx context.Context, email string) (model.User, error) {
	const op = "storage.sqlite.User"

	query := "SELECT id, email, password, created_at FROM users WHERE email = ?"
	row := s.db.QueryRowContext(ctx, query, email)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return model.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, uid int64) (bool, error) {
	const op = "sqlite.IsAdmin"

	query := "SELECT is_admin FROM users WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, uid)

	var app model.App
	err := row.Scan(&app.ID, &app.Name, &app.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (model.App, error) {
	const op = "sqlite.App"

	query := "SELECT id, name, created_at FROM apps WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, appID)

	var app model.App
	err := row.Scan(&app.ID, &app.Name, &app.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return model.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) CreateApp(ctx context.Context, name string) (int64, error) {
	const op = "sqlite.CreateApp"

	res, err := s.db.ExecContext(ctx, "INSERT INTO apps (name) VALUES (?)", name)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAppAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	appID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return appID, nil
}
