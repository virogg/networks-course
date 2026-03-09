package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virogg/networks-course/service/pkg/logger"
)

const (
	defaultMaxConn           = 20
	defaultMinConn           = 5
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = time.Minute

	pingMaxDelay     = 5 * time.Second
	pingInitialDelay = 500 * time.Millisecond
	pingWaitTime     = 30 * time.Second
)

func NewPool(ctx context.Context, dsn string, log logger.Logger, opts ...Option) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	config.MaxConns = defaultMaxConn
	config.MinConns = defaultMinConn
	config.MaxConnLifetime = defaultMaxConnLifetime
	config.MaxConnIdleTime = defaultMaxConnIdleTime
	config.HealthCheckPeriod = defaultHealthCheckPeriod

	for _, opt := range opts {
		opt(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, pingWaitTime)
	defer cancel()

	cnt := 0
	delay := pingInitialDelay
	for {
		cnt++

		if err := pool.Ping(waitCtx); err == nil {
			log.Info("Successfully connected to database")
			return pool, nil
		}

		log.Warn("Database not ready", logger.NewField("attempt", cnt))
		log.Warn("Will retry in", logger.NewField("delay", delay))

		select {
		case <-time.After(delay):
			delay *= 2
			if delay > pingMaxDelay {
				delay = pingMaxDelay
			}
		case <-waitCtx.Done():
			pool.Close()
			return nil, fmt.Errorf("failed to ping database: context done while waiting: %w", waitCtx.Err())
		}
	}
}
