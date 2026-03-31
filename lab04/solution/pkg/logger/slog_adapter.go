package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func New(logLvl string) (*SlogAdapter, error) {
	var log *slog.Logger

	opts := &slog.HandlerOptions{AddSource: true}

	switch logLvl {
	case "local":
		opts.Level = slog.LevelInfo
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	case "dev":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "prod":
		opts.Level = slog.LevelInfo
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	case "test":
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	default: // default value is set to "dev" in config
		opts.Level = slog.LevelDebug
		log = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return &SlogAdapter{logger: log}, nil
}

func (s *SlogAdapter) log(level slog.Level, msg string, fields ...Field) {
	if !s.logger.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, log, Info/Debug/...]

	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	for _, f := range fields {
		r.AddAttrs(slog.Any(f.Key, f.Value))
	}

	_ = s.logger.Handler().Handle(context.Background(), r)
}

func (s *SlogAdapter) Debug(msg string, fields ...Field) {
	s.log(slog.LevelDebug, msg, fields...)
}

func (s *SlogAdapter) Info(msg string, fields ...Field) {
	s.log(slog.LevelInfo, msg, fields...)
}

func (s *SlogAdapter) Warn(msg string, fields ...Field) {
	s.log(slog.LevelWarn, msg, fields...)
}

func (s *SlogAdapter) Error(msg string, fields ...Field) {
	s.log(slog.LevelError, msg, fields...)
}

func (s *SlogAdapter) Sync() error {
	return nil
}
