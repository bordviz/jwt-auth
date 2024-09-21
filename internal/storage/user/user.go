package user

import (
	"context"
	"errors"
	"jwt-auth/internal/domain/dto"
	"jwt-auth/internal/domain/models"
	"jwt-auth/internal/lib/logger/sl"
	"jwt-auth/internal/lib/logger/with"
	"jwt-auth/internal/lib/storage/query"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserDB struct {
	log *slog.Logger
}

func NewUserDB(log *slog.Logger) *UserDB {
	return &UserDB{log: log}
}

func (db *UserDB) Create(ctx context.Context, tx pgx.Tx, user dto.User, requestID string) (uuid.UUID, error) {
	const op = "storage.user.Create"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
		INSERT INTO public.user (email)
		VALUES ($1)
		RETURNING id;
	`

	db.log.Debug("create user query", slog.String("query", query.QueryToString(q)))

	var id uuid.UUID
	if err := tx.QueryRow(ctx, q, user.Email).Scan(&id); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			db.log.Debug("duplicate email", slog.String("email", user.Email))
			return uuid.Nil, errors.New("user with this email already exists")
		}
		db.log.Error("failed to create user", sl.Err(err))
		return uuid.Nil, err
	}

	db.log.Debug("new user was successfully created", slog.String("user_id", id.String()))
	return id, nil
}

func (db *UserDB) GetUserByID(ctx context.Context, tx pgx.Tx, id uuid.UUID, requestID string) (models.UserWithRefresh, error) {
	const op = "storage.user.GetUserByID"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
        SELECT id, email, refresh_id
        FROM public.user AS u
		JOIN refresh AS r ON u.id = r.user_id
        WHERE id = $1;
    `

	db.log.Debug("get user by id query", slog.String("query", query.QueryToString(q)))

	var user models.UserWithRefresh
	if err := tx.QueryRow(ctx, q, id).Scan(&user.ID, &user.Email, &user.RefreshID); err != nil {
		if err == pgx.ErrNoRows {
			db.log.Debug("user not found", slog.String("user_id", id.String()))
			return models.UserWithRefresh{}, errors.New("user not found")
		}
		db.log.Error("failed to get user by id", sl.Err(err))
		return models.UserWithRefresh{}, err
	}

	db.log.Debug("user successfully fetched", slog.String("user_id", id.String()))
	return user, nil
}
