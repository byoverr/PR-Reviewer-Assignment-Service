package main

import (
	"log"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/config"
	loggerConstructor "github.com/byoverr/PR-Reviewer-Assignment-Service/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := loggerConstructor.New(cfg.LogLevel)
	logger.Info("level", cfg.LogLevel)

}
