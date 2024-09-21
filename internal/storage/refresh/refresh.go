package refresh

import (
	"context"
	"jwt-auth/internal/lib/logger/sl"
	"jwt-auth/internal/lib/logger/with"
	"jwt-auth/internal/lib/storage/query"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RefreshDB struct {
	log *slog.Logger
}

func NewRefreshDB(log *slog.Logger) *RefreshDB {
	return &RefreshDB{
		log: log,
	}
}

func (db *RefreshDB) Create(ctx context.Context, tx pgx.Tx, userID uuid.UUID, requestID string) (int, error) {
	const op = "storage.refresh.Create"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
		INSERT INTO refresh (user_id)
		VALUES ($1)
		RETURNING refresh_id;
	`

	db.log.Debug("create refresh id query", slog.String("query", query.QueryToString(q)))

	var id int
	if err := tx.QueryRow(ctx, q, userID).Scan(&id); err != nil {
		db.log.Error("failed to create refresh id", sl.Err(err))
		return 0, err
	}

	db.log.Debug("new refresh id successfully created", slog.Int("refresh_id", id))
	return id, nil
}

func (db *RefreshDB) Delete(ctx context.Context, tx pgx.Tx, userID uuid.UUID, requestID string) error {
	const op = "storage.refresh.Delete"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
        DELETE FROM refresh
        WHERE user_id = $1;
    `

	db.log.Debug("delete refresh id query", slog.String("query", query.QueryToString(q)))

	if _, err := tx.Exec(ctx, q, userID); err != nil {
		db.log.Error("failed to delete refresh id", sl.Err(err))
		return err
	}

	db.log.Debug("refresh id successfully deleted", slog.String("user_id", userID.String()))
	return nil
}
