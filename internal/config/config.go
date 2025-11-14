package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBURL    string `yaml:"db_url" env:"DB_URL" env-default:"postgres://user:pass@localhost:5432/prdb?sslmode=disable" env-description:"PostgreSQL connection URL"`
	Port     string `yaml:"port" env:"PORT" env-default:"8080" env-description:"HTTP server port"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info" env-description:"Logging level (debug/info/warn/error)"`
}

func Load() (*Config, error) {
	var cfg Config

	configPath := "./configs/config_example.yml"
	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return nil, err
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
