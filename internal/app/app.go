package app

import (
	"log/slog"

	grpcapp "github.com/JSONStatham/sso/internal/app/grpc"
	"github.com/JSONStatham/sso/internal/config"
	"github.com/JSONStatham/sso/internal/services/auth"
	"github.com/JSONStatham/sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
	Storage *sqlite.Storage
}

func New(log *slog.Logger, cfg *config.Config) *App {
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, cfg, storage)
	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{GRPCSrv: grpcApp, Storage: storage}
}
