package main

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/virogg/networks-course/service/internal/application"
	"github.com/virogg/networks-course/service/internal/infrastructure/config"
	"github.com/virogg/networks-course/service/pkg/logger"
)

func main() {
	_ = godotenv.Load()
	cfg := config.MustLoad()
	log := logger.Must(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := application.Must(ctx, cfg, log)
	defer app.Shutdown()

	log.Info("Starting service")

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic", logger.NewField("recover", r), logger.NewField("stack", string(debug.Stack())))
			os.Exit(1)
		}
	}()

	app.MustRun(ctx)

	log.Info("Service stopped successfully")
}
