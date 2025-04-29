package db

import (
	"context"

	"github.com/dvvnFrtn/capstone-backend/config"
	"github.com/jackc/pgx/v5"
)

func NewPostgreConn(ctx context.Context, cfg *config.DB) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, cfg.DSN())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
