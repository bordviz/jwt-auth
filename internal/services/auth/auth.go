package auth

import (
	"context"
	"errors"
	"jwt-auth/internal/config"
	"jwt-auth/internal/domain/dto"
	"jwt-auth/internal/domain/models"
	"jwt-auth/internal/lib/jwt"
	"jwt-auth/internal/lib/logger/sl"
	"jwt-auth/internal/lib/logger/with"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService struct {
	log       *slog.Logger
	pool      *pgxpool.Pool
	cfg       config.JWT
	userDB    UserDB
	refreshDB RefreshDB
}

type UserDB interface {
	Create(ctx context.Context, tx pgx.Tx, user dto.User, requestID string) (uuid.UUID, error)
	GetUserByID(ctx context.Context, tx pgx.Tx, id uuid.UUID, requestID string) (models.UserWithRefresh, error)
}

type RefreshDB interface {
	Create(ctx context.Context, tx pgx.Tx, userID uuid.UUID, requestID string) (int, error)
	Delete(ctx context.Context, tx pgx.Tx, userID uuid.UUID, requestID string) error
}

func NewAuthService(log *slog.Logger, pool *pgxpool.Pool, cfg config.JWT, userDB UserDB, refreshDB RefreshDB) *AuthService {
	return &AuthService{
		log:       log,
		pool:      pool,
		cfg:       cfg,
		userDB:    userDB,
		refreshDB: refreshDB,
	}
}

func (s *AuthService) CreateUser(ctx context.Context, user dto.User, ipAddress string, requestID string) (uuid.UUID, error) {
	const op = "services.auth.CreateUser"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	userID, err := s.userDB.Create(ctx, tx, user, requestID)
	if err != nil {
		s.log.Error("failed to create user", sl.Err(err))
		return uuid.Nil, err
	}
	if _, err = s.refreshDB.Create(ctx, tx, userID, requestID); err != nil {
		s.log.Error("failed to create refresh id", sl.Err(err))
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err))
		return uuid.Nil, err
	}

	s.log.Debug("user successfully created")
	return userID, nil
}

func (s *AuthService) GetTokens(ctx context.Context, userID uuid.UUID, ipAddress string, requestID string) (models.TokensPair, error) {
	const op = "services.auth.GetTokens"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return models.TokensPair{}, err
	}
	defer tx.Rollback(ctx)

	user, err := s.userDB.GetUserByID(ctx, tx, userID, requestID)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return models.TokensPair{}, err
	}

	accT, err := jwt.CreateToken(
		user.ID,
		ipAddress,
		user.RefreshID,
		s.cfg.AccessSecret,
		s.cfg.AccessTokenLifetime,
	)
	if err != nil {
		s.log.Error("failed to create access token", sl.Err(err))
		return models.TokensPair{}, err
	}
	refT, err := jwt.CreateToken(
		user.ID,
		ipAddress,
		user.RefreshID,
		s.cfg.RefreshSecret,
		s.cfg.RefreshTokenLifetime,
	)
	if err != nil {
		s.log.Error("failed to create refresh token", sl.Err(err))
		return models.TokensPair{}, err
	}

	tokenPair := models.TokensPair{
		AccessToken:  accT,
		RefreshToken: refT,
	}

	return tokenPair, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string, ipAddress string, requestID string) (models.TokensPair, error) {
	const op = "services.auth.RefreshTokens"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	decodeToken, err := jwt.DecodeToken(refreshToken, s.cfg.RefreshSecret)
	if err != nil {
		s.log.Error("failed to decode token", sl.OpErr(op, err))
		return models.TokensPair{}, errors.New("unauthorized")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return models.TokensPair{}, err
	}
	defer tx.Rollback(ctx)

	user, err := s.userDB.GetUserByID(ctx, tx, decodeToken.UserID, requestID)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return models.TokensPair{}, err
	}

	if decodeToken.IpAddress != ipAddress {
		s.log.Error("unknown ip address", sl.Err(err))
	}

	if decodeToken.RefreshID != user.RefreshID {
		s.log.Error("invalid refresh token")
		return models.TokensPair{}, errors.New("unauthorized")
	}

	if err := s.refreshDB.Delete(ctx, tx, user.ID, requestID); err != nil {
		s.log.Error("failed to delete an old refresh id")
		return models.TokensPair{}, err
	}
	newRefreshID, err := s.refreshDB.Create(ctx, tx, user.ID, requestID)
	if err != nil {
		s.log.Error("failed to create new refresh id", sl.Err(err))
		return models.TokensPair{}, err
	}

	accT, err := jwt.CreateToken(
		user.ID,
		ipAddress,
		newRefreshID,
		s.cfg.AccessSecret,
		s.cfg.AccessTokenLifetime,
	)
	if err != nil {
		s.log.Error("failed to create access token", sl.Err(err))
		return models.TokensPair{}, err
	}
	refT, err := jwt.CreateToken(
		user.ID,
		ipAddress,
		newRefreshID,
		s.cfg.RefreshSecret,
		s.cfg.RefreshTokenLifetime,
	)
	if err != nil {
		s.log.Error("failed to create refresh token", sl.Err(err))
		return models.TokensPair{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err))
		return models.TokensPair{}, err
	}

	tokenPair := models.TokensPair{
		AccessToken:  accT,
		RefreshToken: refT,
	}

	return tokenPair, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, token string, requestID string) (models.User, error) {
	const op = "services.auth.GetCurrentUser"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	decodeToken, err := jwt.DecodeToken(token, s.cfg.AccessSecret)
	if err != nil {
		s.log.Error("failed to decode token", sl.OpErr(op, err))
		return models.User{}, errors.New("unauthorized")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return models.User{}, err
	}
	defer tx.Rollback(ctx)

	user, err := s.userDB.GetUserByID(ctx, tx, decodeToken.UserID, requestID)
	if err != nil {
		s.log.Error("failed to get user", sl.Err(err))
		return models.User{}, err
	}

	return models.User{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}
