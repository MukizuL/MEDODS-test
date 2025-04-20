package app

import (
	"MEDODS-test/internal/config"
	"MEDODS-test/internal/storage/pgstorage"
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"os/signal"
	"path/filepath"
	"syscall"
)

type repo interface {
	SaveRefreshToken(ctx context.Context, refresh, session string) error
	CheckoutRefreshToken(ctx context.Context, refresh, session string) error
}

type Application struct {
	storage repo
	logger  *zap.Logger
	params  *config.Params
}

func NewApplication(storage repo, logger *zap.Logger, params *config.Params) *Application {
	return &Application{
		storage: storage,
		logger:  logger,
		params:  params,
	}
}

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	params := config.GetParams()

	err = Migrate(params.DSN)
	if err != nil {
		return err
	}

	db, err := pgstorage.New(ctx, params.DSN, log)
	if err != nil {
		return err
	}
	defer db.Close()

	app := NewApplication(db, log, params)

	r := NewRouter(app)

	err = runServer(ctx, params.Addr, r)
	if err != nil {
		return err
	}

	return nil
}

func Migrate(DSN string) error {
	_, err := filepath.Abs("./migrations")
	if err != nil {
		return err
	}

	m, err := migrate.New("file://migrations", DSN+"?sslmode=disable")
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
