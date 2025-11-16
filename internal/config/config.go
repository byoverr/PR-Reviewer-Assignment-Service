package config

import (
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DBURL       string `env:"DB_URL"        env-required:"true" env-description:"PostgreSQL"`
	Port        string `env:"PORT"                              env-description:"HTTP server port"                   env-default:"8080"`
	LogLevel    string `env:"LOG_LEVEL"                         env-description:"Logging level"                      env-default:"info"`
	LogOutput   string `env:"LOG_OUTPUT"                        env-description:"Log output: stdout or file"         env-default:"stdout"`
	LogFilePath string `env:"LOG_FILE_PATH"                     env-description:"Log file path (if LOG_OUTPUT=file)" env-default:"./app.log"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		//nolint: sloglint // Using default logger for config loading is acceptable
		slog.Info("No .env file found, relying on environment variables and defaults")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
