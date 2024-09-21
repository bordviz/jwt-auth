package postgresql

import (
	"context"
	"fmt"
	"jwt-auth/internal/config"
	"jwt-auth/internal/lib/logger/sl"
	"jwt-auth/internal/lib/storage/repeateble"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConection(ctx context.Context, log *slog.Logger, cfg config.Database) (*pgxpool.Pool, error) {
	const op = "database.postgresql.NewClient"

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	var pool *pgxpool.Pool

	err := repeateble.DoWithTries(func() error {
		log.Info("database connection attempt")
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()

		var connErr error
		pool, connErr = pgxpool.New(ctx, dsn)

		if connErr != nil {
			log.Error("failed database connection", sl.OpErr(op, connErr))
			return connErr
		}

		if err := pool.Ping(ctx); err != nil {
			log.Error("database conection failed", sl.OpErr(op, err))
			return err
		}

		return nil
	}, cfg.Attempts, cfg.Delay)

	if err != nil {
		log.Error("failed connect to database", sl.OpErr(op, err))
		return nil, err
	}

	log.Info("database connection established", slog.String("op", op))

	return pool, nil
}
