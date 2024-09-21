package with

import "log/slog"

func WithOpAndRequestID(log *slog.Logger, op string, request_id string) *slog.Logger {
	return log.With(
		slog.String("op", op),
		slog.String("request_id", request_id),
	)
}
