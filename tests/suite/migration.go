package suite

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(t *testing.T) {
	t.Helper()

	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	migrationsPath := filepath.Join(rootDir, "migrations")

	if _, err := os.Stat(migrationsPath); err != nil {
		t.Errorf("migrations path error: %v", err)
	}

	m, err := migrate.New(
		"file://"+filepath.ToSlash(migrationsPath),
		"sqlite3://file::memory:?cache=shared",
	)
	if err != nil {
		t.Error(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			t.Error("no migrations to apply")
			return
		}

		t.Error(err)
	}
}
