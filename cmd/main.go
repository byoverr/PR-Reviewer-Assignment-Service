package main

import (
	"context"
	"log"
	"os"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/config"
	loggerConstructor "github.com/byoverr/PR-Reviewer-Assignment-Service/internal/logger"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := loggerConstructor.New(cfg.LogLevel)
	logger.Info("level", cfg.LogLevel)

	// DB pool
	db, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		logger.Error("failed to connect to DB", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repository
	//userRepo := repository.NewUserRepo(db)
	teamRepo := repository.NewTeamRepo(db)
	//prRepo := repository.NewPRRepo(db)

	team := &models.Team{
		Name: "John Team",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}
	err = teamRepo.CreateTeam(context.Background(), team)
	logger.Info("ff", err)

	// Services

	// TODO: Services

	// Handlers
	// TODO: handlers

	// Gin
	// TODO: gin

	//router := gin.Default()
	//router.GET("/", func(c *gin.Context) {
	//	time.Sleep(5 * time.Second)
	//	c.String(http.StatusOK, "Welcome Gin Server")
	//})
	//
	//srv := &http.Server{
	//	Addr:    ":8080",
	//	Handler: router.Handler(),
	//}
	//
	//go func() {
	//	// service connections
	//	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//		log.Fatalf("listen: %s\\n", err)
	//	}
	//}()
	//
	//// Wait for interrupt signal to gracefully shutdown the server with
	//// a timeout of 5 seconds.
	//quit := make(chan os.Signal, 1)
	//// kill (no param) default send syscall.SIGTERM
	//// kill -2 is syscall.SIGINT
	//// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//<-quit
	//log.Println("Shutdown Server ...")
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//if err := srv.Shutdown(ctx); err != nil {
	//	log.Fatal("Server Shutdown:", err)
	//}
	//// catching ctx.Done(). timeout of 5 seconds.
	//select {
	//case <-ctx.Done():
	//	log.Println("timeout of 5 seconds.")
	//}
	//log.Println("Server exiting")
}
