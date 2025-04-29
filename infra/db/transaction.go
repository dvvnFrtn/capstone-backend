package db

import (
	"context"
	"fmt"
	"log/slog"

	database "github.com/dvvnFrtn/capstone-backend/infra/db/sqlc"
	"github.com/dvvnFrtn/capstone-backend/internal/types"
	"github.com/dvvnFrtn/capstone-backend/pkg/errs"
	"github.com/jackc/pgx/v5"
)

func RunTransaction(ctx context.Context, conn *pgx.Conn, cb func(q *database.Queries) error) error {
	const op errs.Op = "db.RunTransaction"
	reqID := ctx.Value(types.RequestIDKey)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errs.New(
			op,
			errs.Internal,
			fmt.Errorf("failed to start tx: %w", err),
		)
	} else {
		slog.Info("starting database tx", "request_id", reqID)
	}

	queries := database.New(conn)
	qtx := queries.WithTx(tx)
	if err := cb(qtx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errs.New(
				op,
				errs.Internal,
				fmt.Errorf("failed to rollback tx: %w", err),
			)
		} else {
			slog.Info("rollback database tx", "request_id", reqID)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.New(
			op,
			errs.Internal,
			fmt.Errorf("failed to commit tx: %w", err),
		)
	} else {
		slog.Info("commiting database tx", "request_id", reqID)
	}

	return nil
}
