package config

import (
	"fmt"
	"os"
	"time"

	pkgconfig "github.com/virogg/networks-course/service/pkg/config"
)

type Config struct {
	// log:
	LogLevel string
	// db:
	DB DatabaseConfig
	// server:
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	// storage:
	ImageDir string
}

func Load() (*Config, error) {
	logLevel := pkgconfig.GetEnvWithDefault("LOG_LEVEL", "dev")

	db, err := LoadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	port := pkgconfig.GetEnvWithDefault("PORT", "8080")
	imageDir := pkgconfig.GetEnvWithDefault("IMAGE_DIR", "/tmp/images")

	return &Config{
		LogLevel:        logLevel,
		DB:              *db,
		Port:            port,
		ReadTimeout:     pkgconfig.GetEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    pkgconfig.GetEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     pkgconfig.GetEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: pkgconfig.GetEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		ImageDir:        imageDir,
	}, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
