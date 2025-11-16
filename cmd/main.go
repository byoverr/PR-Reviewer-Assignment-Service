package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/config"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/handlers"
	loggerConstructor "github.com/byoverr/PR-Reviewer-Assignment-Service/internal/logger"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := loggerConstructor.New(cfg.LogLevel, cfg.LogOutput, cfg.LogFilePath)

	ctx := context.Background()
	poolCfg, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		logger.Error("failed to parse DB config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error("failed to connect to DB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	teamRepo := repository.NewTeamRepo(db)
	userRepo := repository.NewUserRepo(db)
	prRepo := repository.NewPRRepo(db)

	// Services
	teamSvc := services.NewTeamService(teamRepo, userRepo, logger)
	userSvc := services.NewUserService(userRepo, prRepo, logger)
	prSvc := services.NewPRService(prRepo, userRepo, logger)

	// Handlers
	teamHandler := handlers.NewTeamHandler(teamSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	prHandler := handlers.NewPRHandler(prSvc, logger)

	// Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Middleware (recovery for panics)
	router.Use(gin.Recovery())

	// Routes
	handlers.SetupRoutes(router, prHandler, teamHandler, userHandler)

	// Server
	const shutdownTimeout = 5 * time.Second

	const readHeaderTimeout = 60 * time.Second

	srv := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     router,
		ReadTimeout: readHeaderTimeout,
	}

	logger.Info("starting server", slog.String("port", cfg.Port))

	// Grateful shtdown
	go func() {
		if listenErr := srv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			logger.Error("listen: %s\n", slog.String("error", listenErr.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if shutdownErr := srv.Shutdown(ctxShutdown); shutdownErr != nil {
		logger.Error("server forced to shutdown", slog.String("error", shutdownErr.Error()))
	}

	<-ctxShutdown.Done()
	logger.Info("timeout of 5 seconds, server exiting")
	logger.Info("server exiting")
}
