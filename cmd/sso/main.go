package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/JSONStatham/sso/internal/app"
	"github.com/JSONStatham/sso/internal/config"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	app := app.New(log, cfg)

	go app.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	s := <-stop

	log.Info("received signal", slog.String("signal", s.String()))

	app.GRPCSrv.Stop()
	app.Storage.Close()

	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
