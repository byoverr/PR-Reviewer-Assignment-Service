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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepoForUserHandler struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoForUserHandler) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForUserHandler) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForUserHandler) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepoForUserHandler) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoForUserHandler) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepoForUserHandler) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

type mockPRRepoForUserHandler struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepoForUserHandler) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForUserHandler) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *mockPRRepoForUserHandler) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForUserHandler) MergePR(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPRRepoForUserHandler) GetPRsForUser(
	ctx context.Context, userID string,
) ([]models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequestShort), args.Error(1)
}

func (m *mockPRRepoForUserHandler) ExistsPR(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepoForUserHandler) GetTotalPRs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockPRRepoForUserHandler) GetPrsByStatus(ctx context.Context) (int, int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForUserHandler) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForUserHandler) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForUserHandler) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForUserHandler) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForUserHandler) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func TestUserHandler_SetUserActive(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mUserRepo := &mockUserRepoForUserHandler{}
	mPRRepo := &mockPRRepoForUserHandler{}
	svc := services.NewUserService(mUserRepo, mPRRepo, log)
	handler := handlers.NewUserHandler(svc, log)

	t.Run("Success_Activate", func(t *testing.T) {
		reqBody := map[string]any{
			"user_id":   "u1",
			"is_active": true,
		}
		mUserRepo.On("UpdateUserActive", mock.Anything, "u1", true).Return(nil)
		user := &models.User{ID: "u1", Name: "User1", IsActive: true}
		mUserRepo.On("GetUserByID", mock.Anything, "u1").Return(user, nil)

		router := setupRouter()
		router.POST("/users/setIsActive", handler.SetUserActive)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]any{"invalid": "data"}

		router := setupRouter()
		router.POST("/users/setIsActive", handler.SetUserActive)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_GetPRsForUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mUserRepo := &mockUserRepoForUserHandler{}
	mPRRepo := &mockPRRepoForUserHandler{}
	svc := services.NewUserService(mUserRepo, mPRRepo, log)
	handler := handlers.NewUserHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		prs := []models.PullRequestShort{
			{ID: "pr-1", Title: "PR1", Status: "OPEN"},
		}
		mPRRepo.On("GetPRsForUser", mock.Anything, "u1").Return(prs, nil)

		router := setupRouter()
		router.GET("/users/getReview", handler.GetPRsForUser)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mPRRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_MissingUserID", func(t *testing.T) {
		router := setupRouter()
		router.GET("/users/getReview", handler.GetPRsForUser)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
