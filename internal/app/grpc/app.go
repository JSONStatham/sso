package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/JSONStatham/sso/internal/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log    *slog.Logger
	server *grpc.Server
	port   int
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	server := grpc.NewServer()
	authgrpc.Register(server, authService)

	return &App{log: log, server: server, port: port}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running", slog.With("addr", lis.Addr().String()))

	if err := a.server.Serve(lis); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server")

	a.server.GracefulStop()
}
