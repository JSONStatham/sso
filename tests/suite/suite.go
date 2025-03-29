package suite

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	ssov1 "github.com/JSONStatham/protos/gen/go/sso"
	"github.com/JSONStatham/sso/internal/app"
	"github.com/JSONStatham/sso/internal/config"
	slogdiscard "github.com/JSONStatham/sso/internal/utils/logger/sl/handlers"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost           = "localhost"
	defaultWaitTime    = 5 * time.Second
	defaultJWTSecret   = "test-secret-1234567890"
	defaultTestAppName = "test-app"
)

var (
	testAppID int64
	setupOnce sync.Once
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	Log        *slog.Logger
	App        *app.App
	AuthClient ssov1.AuthClient
	GRPCClient *grpc.ClientConn
}

// GetTestAppID returns the ID of the test app created during setup
func (s *Suite) GetTestAppID() int64 {
	return testAppID
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	setupTestEnv(t)
	cfg := config.MustLoadByPath("../config/test.yaml")
	log := slogdiscard.NewDiscardLogger()

	// Initialize application
	app := app.New(log, cfg)

	// Start services in background
	go func() {
		if err := app.GRPCSrv.Run(); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	clientConn := setupGRPCClient(t, cfg)

	setupOnce.Do(func() {
		runMigrations(t)
		createTestApp(t, app)
	})

	t.Cleanup(func() {
		if err := clientConn.Close(); err != nil {
			t.Logf("Failed to close gRPC client: %v", err)
		}
		app.GRPCSrv.Stop()
		app.Storage.Close()

		// Clean environment
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("ENV")
	})

	return context.Background(), &Suite{
		T:          t,
		Cfg:        cfg,
		Log:        log,
		App:        app,
		AuthClient: ssov1.NewAuthClient(clientConn),
		GRPCClient: clientConn,
	}
}

// Helper functions
func setupTestEnv(t *testing.T) {
	t.Helper()

	os.Setenv("ENV", "test")
	os.Setenv("JWT_SECRET", defaultJWTSecret)
}

func createTestApp(t *testing.T, app *app.App) {
	// Create test app
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := app.Storage.CreateApp(ctx, defaultTestAppName)
	require.NoError(t, err)
	testAppID = id
}

func setupGRPCClient(t *testing.T, cfg *config.Config) *grpc.ClientConn {
	t.Helper()

	// Wait for server to be ready
	waitForServerReady(t, cfg.GRPC.Port)

	// Create client
	conn, err := grpc.NewClient(
		net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatalf("failed to create grpc client: %v", err)
	}

	require.NoError(t, err)

	return conn
}

func registerCleanup(t *testing.T, app *app.App, conn *grpc.ClientConn) {
	t.Helper()

}

func waitForServerReady(t *testing.T, port int) {
	t.Helper()

	deadline := time.Now().Add(defaultWaitTime)
	address := net.JoinHostPort(grpcHost, strconv.Itoa(port))

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Server didn't become ready within %v", defaultWaitTime)
}
