package testutil

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/dvvnFrtn/capstone-backend/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewTestContainer(ctx context.Context, cfg *config.DB) (*postgres.PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(cfg.Name),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Pass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		))
	if err != nil {
		return container, err
	}

	return container, nil
}

func MigrateDatabase(dsn string) error {
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("failed to get path caller")
	}

	dir := filepath.Dir(path)
	mDir := filepath.Join(dir, "..", "..", "infra/db/migrations")
	mSrc := fmt.Sprintf("file:%s", mDir)

	m, err := migrate.New(mSrc, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func SetupTestDatabase(ctx context.Context, cfg *config.DB) (*postgres.PostgresContainer, error) {
	container, err := NewTestContainer(ctx, cfg)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, nat.Port("5432"))
	if err != nil {
		return nil, err
	}
	cfg.Port = port.Port()

	if err := MigrateDatabase(cfg.DSN()); err != nil {
		return nil, err
	}

	if err := container.Snapshot(ctx); err != nil {
		return nil, err
	}

	return container, nil
}
