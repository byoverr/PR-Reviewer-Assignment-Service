package e2e_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/handlers"
	loggerConstructor "github.com/byoverr/PR-Reviewer-Assignment-Service/internal/logger"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupE2ETest(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	ctx := context.Background()
	dsn := "postgres://user:pass@localhost:5433/pr_service_e2e?sslmode=disable"

	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()

	migrationsDir := filepath.Join("..", "..", "migrations")
	err = goose.Up(sqlDB, migrationsDir)
	require.NoError(t, err, "failed to run migrations")

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `TRUNCATE TABLE pull_requests, users, teams CASCADE`)
	require.NoError(t, err)

	teamRepo := repository.NewTeamRepo(db)
	userRepo := repository.NewUserRepo(db)
	prRepo := repository.NewPRRepo(db)

	logger := loggerConstructor.New("info", "stdout", "")
	teamSvc := services.NewTeamService(teamRepo, userRepo, logger)
	userSvc := services.NewUserService(userRepo, prRepo, logger)
	prSvc := services.NewPRService(prRepo, userRepo, logger)

	teamHandler := handlers.NewTeamHandler(teamSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	prHandler := handlers.NewPRHandler(prSvc, logger)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	handlers.SetupRoutes(router, prHandler, teamHandler, userHandler)

	return router, db
}

func TestE2E_CreateTeamAndPR(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	t.Run("CreateTeam", func(t *testing.T) {
		team := models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
				{UserID: "u2", Username: "User2", IsActive: true},
				{UserID: "u3", Username: "User3", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreatePR", func(t *testing.T) {
		pr := models.PullRequest{
			ID:       "pr-1",
			Title:    "Test PR",
			AuthorID: "u1",
		}

		body, _ := json.Marshal(pr)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.NotNil(t, response["pr"])
	})
}

func TestE2E_ReassignReviewer(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true),
		('u2', 'User2', 'team1', true),
		('u3', 'User3', 'team1', true)`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers) 
		VALUES ($1, $2, $3, $4, $5)`,
		"pr-1", "Test PR", "u1", "OPEN", []string{"u2"})
	require.NoError(t, err)

	t.Run("ReassignReviewer", func(t *testing.T) {
		reqBody := map[string]string{
			"pull_request_id": "pr-1",
			"old_reviewer_id": "u2",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.NotNil(t, response["pr"])
		assert.NotNil(t, response["replaced_by"])
	})
}

func TestE2E_MergePR(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true)`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status) 
		VALUES ($1, $2, $3, $4)`,
		"pr-2", "Test PR", "u1", "OPEN")
	require.NoError(t, err)

	t.Run("MergePR", func(t *testing.T) {
		reqBody := map[string]string{
			"pull_request_id": "pr-2",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.NotNil(t, response["pr"])

		prData, ok := response["pr"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "MERGED", prData["status"])
	})
}

func TestE2E_GetStats(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true),
		('u2', 'User2', 'team1', true)`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers) 
		VALUES 
		('pr-1', 'PR1', 'u1', 'OPEN', ARRAY['u2']),
		('pr-2', 'PR2', 'u1', 'MERGED', ARRAY['u2'])`)
	require.NoError(t, err)

	t.Run("GetTotalPRs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/prs-total", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.PrsTotal
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.GreaterOrEqual(t, response.TotalPRs, 2)
	})

	t.Run("GetPrsByStatus", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/prs-status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.PrsStatus
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.GreaterOrEqual(t, response.OpenPRs, 1)
		assert.GreaterOrEqual(t, response.MergedPRs, 1)
	})

	t.Run("GetTopReviewers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/top-reviewers", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.TopReviewers
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.NotNil(t, response.TopReviewers)
	})
}

func TestE2E_SetUserActive(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true)`)
	require.NoError(t, err)

	t.Run("SetUserActive_False", func(t *testing.T) {
		reqBody := map[string]any{
			"user_id":   "u1",
			"is_active": false,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.NotNil(t, response["user"])

		userData, ok := response["user"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, false, userData["is_active"])
	})
}

func TestE2E_GetPRsForUser(t *testing.T) {
	router, db := setupE2ETest(t)
	defer db.Close()

	ctx := context.Background()
	_, err := db.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team1")
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true),
		('u2', 'User2', 'team1', true)`)
	require.NoError(t, err)

	_, err = db.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers) 
		VALUES 
		('pr-1', 'PR1', 'u1', 'OPEN', ARRAY['u2']),
		('pr-2', 'PR2', 'u1', 'OPEN', ARRAY['u2'])`)
	require.NoError(t, err)

	t.Run("GetPRsForUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		unmarshalErr := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, unmarshalErr)
		assert.Equal(t, "u2", response["user_id"])
		assert.NotNil(t, response["pull_requests"])

		prs := response["pull_requests"].([]any)
		assert.GreaterOrEqual(t, len(prs), 2)
	})
}
