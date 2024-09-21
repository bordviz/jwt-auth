package logger

import (
	"fmt"
	"jwt-auth/internal/lib/logger/slogpretty"
	"log/slog"
	"os"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

func New(env string) *slog.Logger {
	switch env {
	case EnvLocal:
		return SetupPrettyLogger()
	case EnvDev:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelDebug},
			),
		)
	case EnvProd:
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{Level: slog.LevelInfo},
			),
		)
	default:
		fmt.Printf("unknown env mode: %s, avalible modes: local, dev, prod", env)
		os.Exit(1)
		return nil
	}
}

func SetupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOptions: &slog.HandlerOptions{Level: slog.LevelDebug},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
