package migrations

import (
	"errors"
	"fmt"
	"jwt-auth/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func CreateMigrations(cfg *config.Config, action string) error {
	dsn := createDSN(
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", cfg.MigrationsPath),
		dsn,
	)

	if err != nil {
		return err
	}

	switch action {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return nil
			}
			return err
		}
		return nil
	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return nil
			}
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid migration action: %s", action)
	}
}

func createDSN(user string, pass string, host string, port int, name string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=migrations",
		user,
		pass,
		host,
		port,
		name,
	)
}
