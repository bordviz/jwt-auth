package handlers

import (
	"context"
	"jwt-auth/internal/domain/dto"
	"jwt-auth/internal/domain/models"
	"jwt-auth/internal/lib/logger/sl"
	"jwt-auth/internal/lib/logger/with"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type Handler struct {
	log     *slog.Logger
	service AuthService
}

type AuthService interface {
	CreateUser(ctx context.Context, user dto.User, ipAddress string, requestID string) (uuid.UUID, error)
	GetTokens(ctx context.Context, userID uuid.UUID, ipAddress string, requestID string) (models.TokensPair, error)
	RefreshTokens(ctx context.Context, refreshToken string, ipAddress string, requestID string) (models.TokensPair, error)
	GetCurrentUser(ctx context.Context, token string, requestID string) (models.User, error)
}

func NewHandler(log *slog.Logger, service AuthService) *Handler {
	return &Handler{log: log, service: service}
}

func AddAuthHandlers(log *slog.Logger, service AuthService) func(r chi.Router) {
	handler := NewHandler(log, service)

	ctx := context.TODO()

	return func(r chi.Router) {
		r.Post("/create", handler.CreateUser(ctx))
		r.Get("/tokens/{id}", handler.GetTokens(ctx))
		r.Get("/refresh-tokens", handler.RefreshTokens(ctx))
		r.Get("/current-user", handler.GetCurrentUser(ctx))
	}
}

func (h *Handler) CreateUser(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.CreateUser"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		var user dto.User
		if err := render.Decode(r, &user); err != nil {
			h.log.Error("failed to decode model", sl.Err(err))
			ErrorResponse(w, r, 400, "Bad request")
			return
		}
		if err := user.Validate(); err != nil {
			h.log.Error("validate model error", sl.Err(err))
			ErrorResponse(w, r, 422, err.Error())
			return
		}

		userID, err := h.service.CreateUser(ctx, user, r.RemoteAddr, requestID)
		if err != nil {
			h.log.Error("failed to create user", sl.Err(err))
			ErrorResponse(w, r, 400, err.Error())
			return
		}

		SuccessResponse(w, r, 201, map[string]string{
			"detail": "new user was successfully created",
			"id":     userID.String(),
		})
	}
}

func (h *Handler) GetTokens(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.GetTokens"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			h.log.Error("failed to parse user id", sl.Err(err))
			ErrorResponse(w, r, 400, "Bad request")
			return
		}

		tokens, err := h.service.GetTokens(ctx, userID, r.RemoteAddr, requestID)
		if err != nil {
			h.log.Error("failed to get tokens", sl.Err(err))
			ErrorResponse(w, r, 400, err.Error())
			return
		}

		SuccessResponse(w, r, 200, tokens)
	}
}

func (h *Handler) RefreshTokens(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.RefreshTokens"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		refreshToken := r.Header.Get("Authorization")
		refreshToken = strings.ReplaceAll(refreshToken, "Bearer ", "")
		if refreshToken == "" {
			h.log.Error("missing refresh token")
			ErrorResponse(w, r, 400, "refresh token is required")
			return
		}

		tokens, err := h.service.RefreshTokens(ctx, refreshToken, r.RemoteAddr, requestID)
		if err != nil {
			h.log.Error("failed to refresh tokens", sl.Err(err))
			ErrorResponse(w, r, 401, "Unauthorized")
			return
		}

		SuccessResponse(w, r, 200, tokens)
	}
}

func (h *Handler) GetCurrentUser(ctx context.Context) http.HandlerFunc {
	const op = "handlers.auth.GetCurrentUser"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		token := r.Header.Get("Authorization")
		token = strings.ReplaceAll(token, "Bearer ", "")
		if token == "" {
			h.log.Error("missing access token")
			ErrorResponse(w, r, 401, "access token is required")
			return
		}

		user, err := h.service.GetCurrentUser(ctx, token, requestID)
		if err != nil {
			h.log.Error("failed to get current user", sl.Err(err))
			ErrorResponse(w, r, 401, "Unauthorized")
			return
		}

		SuccessResponse(w, r, 200, user)
	}
}
