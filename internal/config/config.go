package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `yaml:"env" env-required:"true"`
	MigrationsPath string `yaml:"migrations_path" env-required:"true"`
	JWT            `yaml:"jwt" env-required:"true"`
	Database       `yaml:"database" env-required:"true"`
	HTTPServer     `yaml:"http_server" env-required:"true"`
}

type JWT struct {
	AccessTokenLifetime  time.Duration `yaml:"access_token_lifetime" env-required:"true"`
	RefreshTokenLifetime time.Duration `yaml:"refresh_token_lifetime" env-required:"true"`
	AccessSecret         string
	RefreshSecret        string
}

type Database struct {
	Host     string        `yaml:"host" env-required:"true"`
	Port     int           `yaml:"port" env-required:"true"`
	User     string        `yaml:"user" env-required:"true"`
	Name     string        `yaml:"name" env-required:"true"`
	Password string        `yaml:"password" env-required:"true"`
	Timeout  time.Duration `yaml:"timeout" env-required:"true"`
	Attempts int           `yaml:"attempts" env-required:"true"`
	Delay    time.Duration `yaml:"delay" env-required:"true"`
}

type HTTPServer struct {
	Port        int           `yaml:"port" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-reqired:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(".env file not found")
		os.Exit(1)
	}

	configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH"))
	accessSecret := strings.TrimSpace(os.Getenv("ACCESS_SECRET"))
	refreshSecret := strings.TrimSpace(os.Getenv("REFRESH_SECRET"))

	if configPath == "" || accessSecret == "" || refreshSecret == "" {
		fmt.Println("Missing required environment variables in .env file")
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("config file %s not found", configPath)
		os.Exit(1)
	}

	var cfg Config

	cfg.JWT = JWT{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		fmt.Printf("Error reading config file: %s", err.Error())
		os.Exit(1)
	}
	return &cfg

}
