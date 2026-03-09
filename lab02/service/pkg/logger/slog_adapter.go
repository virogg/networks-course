package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func New(logLvl string) (*SlogAdapter, error) {
	var log *slog.Logger

	opts := &slog.HandlerOptions{AddSource: true}

	switch logLvl {
	case "local":
		opts.Level = slog.LevelDebug
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

func Must(logLvl string) *SlogAdapter {
	log, err := New(logLvl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	return log
}

func (s *SlogAdapter) Debug(msg string, fields ...Field) {
	s.logger.Debug(msg, convertFields(fields...)...)
}

func (s *SlogAdapter) Info(msg string, fields ...Field) {
	s.logger.Info(msg, convertFields(fields...)...)
}

func (s *SlogAdapter) Warn(msg string, fields ...Field) {
	s.logger.Warn(msg, convertFields(fields...)...)
}

func (s *SlogAdapter) Error(msg string, fields ...Field) {
	s.logger.Error(msg, convertFields(fields...)...)
}

func (s *SlogAdapter) Sync() error {
	return nil
}

func convertFields(fields ...Field) []any {
	out := make([]any, 0, len(fields))
	for _, f := range fields {
		out = append(out, slog.Any(f.Key, f.Value))
	}
	return out
}
