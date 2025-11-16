package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/handlers"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPRRepoForHandler struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepoForHandler) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForHandler) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *mockPRRepoForHandler) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForHandler) MergePR(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPRRepoForHandler) GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequestShort), args.Error(1)
}

func (m *mockPRRepoForHandler) ExistsPR(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepoForHandler) GetTotalPRs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockPRRepoForHandler) GetPrsByStatus(ctx context.Context) (int, int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForHandler) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForHandler) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForHandler) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForHandler) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForHandler) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

type mockUserRepoForHandler struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoForHandler) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForHandler) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForHandler) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepoForHandler) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoForHandler) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepoForHandler) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestPRHandler_CreatePR(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.PullRequest{
			ID:       "pr-1",
			Title:    "Test PR",
			AuthorID: "u1",
		}

		author := &models.User{ID: "u1", TeamName: "team1"}
		mUserRepo.On("GetUserByID", mock.Anything, "u1").Return(author, nil)
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{
			{ID: "u2", IsActive: true},
		}, nil)
		mPrRepo.On("CreatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(nil)
		reloaded := &models.PullRequest{ID: "pr-1", Status: "OPEN", Reviewers: []string{"u2"}}
		mPrRepo.On("GetPRByID", mock.Anything, "pr-1").Return(reloaded, nil)

		router := setupRouter()
		router.POST("/pullRequest/create", handler.CreatePR)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mUserRepo.AssertExpectations(t)
		mPrRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]any{"invalid": "data"}

		router := setupRouter()
		router.POST("/pullRequest/create", handler.CreatePR)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPRHandler_MergePR(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		reqBody := map[string]string{"pull_request_id": "pr-1"}
		mPrRepo.On("MergePR", mock.Anything, "pr-1").Return(nil)
		pr := &models.PullRequest{ID: "pr-1", Status: "MERGED"}
		mPrRepo.On("GetPRByID", mock.Anything, "pr-1").Return(pr, nil)

		router := setupRouter()
		router.POST("/pullRequest/merge", handler.MergePR)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]string{"invalid": "data"}

		router := setupRouter()
		router.POST("/pullRequest/merge", handler.MergePR)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPRHandler_ReassignReviewer(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		reqBody := map[string]string{
			"pull_request_id": "pr-1",
			"old_reviewer_id": "u2",
		}
		pr := &models.PullRequest{
			ID:        "pr-1",
			Status:    "OPEN",
			Reviewers: []string{"u2"},
			AuthorID:  "u1",
		}
		mPrRepo.On("GetPRByID", mock.Anything, "pr-1").Return(pr, nil)
		oldReviewer := &models.User{ID: "u2", TeamName: "team1"}
		mUserRepo.On("GetUserByID", mock.Anything, "u2").Return(oldReviewer, nil)
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{
			{ID: "u3", IsActive: true},
		}, nil)
		mPrRepo.On("UpdatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(nil)

		router := setupRouter()
		router.POST("/pullRequest/reassign", handler.ReassignReviewer)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
		mUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]string{"invalid": "data"}

		router := setupRouter()
		router.POST("/pullRequest/reassign", handler.ReassignReviewer)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPRHandler_GetTotalPRs(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetTotalPRs", mock.Anything).Return(10, nil)

		router := setupRouter()
		router.GET("/stats/total", handler.GetTotalPRs)

		req := httptest.NewRequest(http.MethodGet, "/stats/total", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetPrsByStatus(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetPrsByStatus", mock.Anything).Return(5, 3, nil)

		router := setupRouter()
		router.GET("/stats/status", handler.GetPrsByStatus)

		req := httptest.NewRequest(http.MethodGet, "/stats/status", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetAssignmentsPerUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		assignments := []models.UserAssignment{
			{UserID: "u1", Name: "User1", Count: 5},
		}
		mPrRepo.On("GetAssignmentsPerUser", mock.Anything).Return(assignments, nil)

		router := setupRouter()
		router.GET("/stats/assignments", handler.GetAssignmentsPerUser)

		req := httptest.NewRequest(http.MethodGet, "/stats/assignments", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetTopReviewers(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		top := []models.UserAssignment{
			{UserID: "u1", Name: "User1", Count: 10},
		}
		mPrRepo.On("GetTopReviewers", mock.Anything).Return(top, nil)

		router := setupRouter()
		router.GET("/stats/top-reviewers", handler.GetTopReviewers)

		req := httptest.NewRequest(http.MethodGet, "/stats/top-reviewers", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetAvgCloseTime(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetAvgCloseTime", mock.Anything).Return(86400.0, 5, nil)

		router := setupRouter()
		router.GET("/stats/avg-close-time", handler.GetAvgCloseTime)

		req := httptest.NewRequest(http.MethodGet, "/stats/avg-close-time", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetIdleUsersPerTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		metrics := []models.TeamMetric{
			{TeamName: "team1", Count: 2},
		}
		mPrRepo.On("GetIdleUsersPerTeam", mock.Anything).Return(metrics, nil)

		router := setupRouter()
		router.GET("/stats/idle-users-per-team", handler.GetIdleUsersPerTeam)

		req := httptest.NewRequest(http.MethodGet, "/stats/idle-users-per-team", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}

func TestPRHandler_GetNeedyPRsPerTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepoForHandler{}
	mUserRepo := &mockUserRepoForHandler{}
	svc := services.NewPRService(mPrRepo, mUserRepo, log)
	handler := handlers.NewPRHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		metrics := []models.TeamMetric{
			{TeamName: "team1", Count: 3},
		}
		mPrRepo.On("GetNeedyPRsPerTeam", mock.Anything).Return(metrics, nil)

		router := setupRouter()
		router.GET("/stats/needy-prs-per-team", handler.GetNeedyPRsPerTeam)

		req := httptest.NewRequest(http.MethodGet, "/stats/needy-prs-per-team", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPrRepo.AssertExpectations(t)
	})
}
