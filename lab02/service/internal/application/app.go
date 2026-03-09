package application

import (
	"context"
	"os"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/virogg/networks-course/service/internal/infrastructure/config"
	pgrepo "github.com/virogg/networks-course/service/internal/infrastructure/repository/postgres/products"
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/http"
	"github.com/virogg/networks-course/service/internal/service/products"
	pkghttp "github.com/virogg/networks-course/service/pkg/http"
	"github.com/virogg/networks-course/service/pkg/logger"
	"github.com/virogg/networks-course/service/pkg/postgres"
)

type App struct {
	server *pkghttp.Server
	pool   *pgxpool.Pool
	log    logger.Logger
}

func New(ctx context.Context, cfg *config.Config, log logger.Logger) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DB.GetDSN(), log)
	if err != nil {
		log.Error("Failed to connect to database", logger.NewField("error", err))
		return nil, err
	}

	c := trmpgx.DefaultCtxGetter
	trManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	productsRepo := pgrepo.NewProductsPostgresRepository(pool, c)

	productsService := products.NewService(trManager, productsRepo, cfg.ImageDir)

	router := http.NewRouter(log, productsService)

	serverConfig := pkghttp.ServerConfig{
		Port:            cfg.Port,
		Handler:         router,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	}

	server := pkghttp.NewServer(log, serverConfig)

	return &App{
		server: server,
		pool:   pool,
		log:    log,
	}, nil
}

func Must(ctx context.Context, cfg *config.Config, log logger.Logger) *App {
	app, err := New(ctx, cfg, log)
	if err != nil {
		log.Error("Failed to initialize application", logger.NewField("error", err))
		os.Exit(1)
	}
	return app
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("Starting application")
	if err := a.server.Start(ctx); err != nil {
		a.log.Error("Failed to start server", logger.NewField("error", err))
		return err
	}
	return nil
}

func (a *App) MustRun(ctx context.Context) {
	if err := a.Run(ctx); err != nil {
		a.log.Error("Application failed to run", logger.NewField("error", err))
		os.Exit(1)
	}
}

func (a *App) Shutdown() {
	a.log.Info("Shutting down application")

	a.log.Info("Closing database connection pool")
	a.pool.Close()

	a.log.Info("Application shutdown complete")
}
