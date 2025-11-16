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

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/handlers"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTeamRepoForHandler struct {
	mock.Mock
	repository.TeamRepository
}

type mockUserRepoForTeamHandler struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockTeamRepoForHandler) CreateTeam(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockUserRepoForTeamHandler) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForTeamHandler) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepoForTeamHandler) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoForTeamHandler) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepoForTeamHandler) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func (m *mockTeamRepoForHandler) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *mockUserRepoForTeamHandler) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestTeamHandler_CreateTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepoForHandler{}
	mUserRepo := &mockUserRepoForTeamHandler{}
	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)
	handler := handlers.NewTeamHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		reqBody := models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
			},
		}
		mTeamRepo.On("CreateTeam", mock.Anything, mock.AnythingOfType("*models.Team")).Return(nil)
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
			},
		}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").Return(team, nil)

		router := setupRouter()
		router.POST("/team/add", handler.CreateTeam)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mTeamRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]any{"invalid": "data"}

		router := setupRouter()
		router.POST("/team/add", handler.CreateTeam)

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTeamHandler_GetTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepoForHandler{}
	mUserRepo := &mockUserRepoForTeamHandler{}
	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)
	handler := handlers.NewTeamHandler(svc, log)

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
			},
		}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").Return(team, nil)

		router := setupRouter()
		router.GET("/team/get", handler.GetTeam)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=team1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mTeamRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_MissingTeamName", func(t *testing.T) {
		router := setupRouter()
		router.GET("/team/get", handler.GetTeam)

		req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTeamHandler_AddMemberToTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepoForHandler{}
	mUserRepo := &mockUserRepoForTeamHandler{}
	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)
	handler := handlers.NewTeamHandler(svc, log)

	router := setupRouter()
	router.POST("/team/add-member", handler.AddMemberToTeam)

	t.Run("InvalidInput", func(t *testing.T) {
		reqBody := map[string]any{"invalid": "data"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add-member", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("TeamNotFound", func(t *testing.T) {
		mTeamRepo.On("GetTeamByName", mock.Anything, "unknown-team").
			Return(nil, apperrors.ErrNotFound)

		reqBody := map[string]any{
			"team_name": "unknown-team",
			"member": map[string]any{
				"user_id":   "user1",
				"username":  "John",
				"is_active": true,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add-member", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)

		mTeamRepo.AssertExpectations(t)
	})

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{Name: "team1"}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").
			Return(team, nil)
		mUserRepo.On("UpsertUser", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(nil)

		reqBody := map[string]any{
			"team_name": "team1",
			"member": map[string]any{
				"user_id":   "user1",
				"username":  "John",
				"is_active": true,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/team/add-member", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "member added successfully", resp["message"])

		mTeamRepo.AssertExpectations(t)
		mUserRepo.AssertExpectations(t)
	})
}
