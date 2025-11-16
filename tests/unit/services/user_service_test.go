package services_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepoForUserService struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoForUserService) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForUserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForUserService) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepoForUserService) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoForUserService) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepoForUserService) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

type mockPRRepoForUserService struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepoForUserService) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForUserService) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *mockPRRepoForUserService) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForUserService) MergePR(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPRRepoForUserService) GetPRsForUser(
	ctx context.Context, userID string,
) ([]models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequestShort), args.Error(1)
}

func (m *mockPRRepoForUserService) ExistsPR(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepoForUserService) GetTotalPRs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockPRRepoForUserService) GetPrsByStatus(ctx context.Context) (int, int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForUserService) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForUserService) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForUserService) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForUserService) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForUserService) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForUserService) GetOpenPRsWithReviewersFromTeam(
	ctx context.Context,
	teamName string,
) ([]models.PullRequest, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func TestUserService_SetUserActive(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mUserRepo := &mockUserRepoForUserService{}
	mPRRepo := &mockPRRepoForUserService{}

	svc := services.NewUserService(mUserRepo, mPRRepo, log)

	t.Run("Success_Activate", func(t *testing.T) {
		mUserRepo.On("UpdateUserActive", mock.Anything, "u1", true).Return(nil)
		user := &models.User{ID: "u1", Name: "User1", IsActive: true}
		mUserRepo.On("GetUserByID", mock.Anything, "u1").Return(user, nil)

		result, err := svc.SetUserActive(context.Background(), "u1", true)
		require.NoError(t, err)
		assert.True(t, result.IsActive)
		mUserRepo.AssertExpectations(t)
	})

	t.Run("Success_Deactivate", func(t *testing.T) {
		mUserRepo2 := &mockUserRepoForUserService{}
		mPRRepo2 := &mockPRRepoForUserService{}
		svc2 := services.NewUserService(mUserRepo2, mPRRepo2, log)

		mUserRepo2.On("UpdateUserActive", mock.Anything, "u1", false).Return(nil)
		user := &models.User{ID: "u1", Name: "User1", IsActive: false}
		mUserRepo2.On("GetUserByID", mock.Anything, "u1").Return(user, nil)

		result, err := svc2.SetUserActive(context.Background(), "u1", false)
		require.NoError(t, err)
		assert.False(t, result.IsActive)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		_, err := svc.SetUserActive(context.Background(), "", true)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mUserRepo3 := &mockUserRepoForUserService{}
		mPRRepo3 := &mockPRRepoForUserService{}
		svc3 := services.NewUserService(mUserRepo3, mPRRepo3, log)

		mUserRepo3.On("UpdateUserActive", mock.Anything, "u-nonexist", true).Return(apperrors.ErrNotFound)

		_, err := svc3.SetUserActive(context.Background(), "u-nonexist", true)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("ReloadFailed", func(t *testing.T) {
		mUserRepo4 := &mockUserRepoForUserService{}
		mPRRepo4 := &mockPRRepoForUserService{}
		svc4 := services.NewUserService(mUserRepo4, mPRRepo4, log)

		mUserRepo4.On("UpdateUserActive", mock.Anything, "u1", true).Return(nil)
		mUserRepo4.On("GetUserByID", mock.Anything, "u1").Return(nil, apperrors.ErrInternal)

		_, err := svc4.SetUserActive(context.Background(), "u1", true)
		assert.ErrorIs(t, err, apperrors.ErrInternal)
	})
}

func TestUserService_GetPRsForUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mUserRepo := &mockUserRepoForUserService{}
	mPRRepo := &mockPRRepoForUserService{}

	svc := services.NewUserService(mUserRepo, mPRRepo, log)

	t.Run("Success", func(t *testing.T) {
		prs := []models.PullRequestShort{
			{ID: "pr-1", Title: "PR1", Status: "OPEN"},
			{ID: "pr-2", Title: "PR2", Status: "OPEN"},
		}
		mPRRepo.On("GetPRsForUser", mock.Anything, "u1").Return(prs, nil)

		result, err := svc.GetPRsForUser(context.Background(), "u1")
		require.NoError(t, err)
		assert.Len(t, result, 2)
		mPRRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		_, err := svc.GetPRsForUser(context.Background(), "")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("Error", func(t *testing.T) {
		mUserRepo5 := &mockUserRepoForUserService{}
		mPRRepo5 := &mockPRRepoForUserService{}
		svc5 := services.NewUserService(mUserRepo5, mPRRepo5, log)

		mPRRepo5.On("GetPRsForUser", mock.Anything, "u1").Return(nil, apperrors.ErrInternal)

		_, err := svc5.GetPRsForUser(context.Background(), "u1")
		assert.Error(t, err)
	})

	t.Run("EmptyList", func(t *testing.T) {
		mUserRepo6 := &mockUserRepoForUserService{}
		mPRRepo6 := &mockPRRepoForUserService{}
		svc6 := services.NewUserService(mUserRepo6, mPRRepo6, log)

		mPRRepo6.On("GetPRsForUser", mock.Anything, "u1").Return([]models.PullRequestShort{}, nil)

		result, err := svc6.GetPRsForUser(context.Background(), "u1")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestUserService_DeactivateUsersByTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Success_NoPRs", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		activeUsers := []models.User{
			{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
			{ID: "u2", Name: "User2", TeamName: "team1", IsActive: true},
		}
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
		mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
		mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return([]models.PullRequest{}, nil)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		require.NoError(t, err)
		mUserRepo.AssertExpectations(t)
		mPRRepo.AssertExpectations(t)
	})

	t.Run("Success_WithPRReassignment_PartialTeam", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		activeUsersBefore := []models.User{
			{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
			{ID: "u2", Name: "User2", TeamName: "team1", IsActive: true},
		}

		activeUsersAfter := []models.User{
			{ID: "u3", Name: "User3", TeamName: "team1", IsActive: true},
		}

		pr := &models.PullRequest{
			ID:        "pr-1",
			Title:     "Test PR",
			AuthorID:  "author1",
			Status:    "OPEN",
			Reviewers: []string{"u1", "u2"},
		}

		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsersBefore, nil).Once()
		mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
		mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return([]models.PullRequest{*pr}, nil)
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsersAfter, nil).Once()

		mPRRepo.On("UpdatePR", mock.Anything, mock.MatchedBy(func(p *models.PullRequest) bool {
			return p.ID == "pr-1" && len(p.Reviewers) >= 1 && p.Reviewers[0] == "u3"
		})).Return(nil)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		require.NoError(t, err)
		mUserRepo.AssertExpectations(t)
		mPRRepo.AssertExpectations(t)
	})

	t.Run("Success_NoActiveReplacement", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		activeUsers := []models.User{
			{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
			{ID: "u2", Name: "User2", TeamName: "team1", IsActive: true},
		}

		activeUsersAfter := []models.User{}

		pr := &models.PullRequest{
			ID:        "pr-1",
			Title:     "Test PR",
			AuthorID:  "author1",
			Status:    "OPEN",
			Reviewers: []string{"u1", "u2"},
		}

		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil).Once()
		mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
		mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return([]models.PullRequest{*pr}, nil)
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsersAfter, nil).Once()

		mPRRepo.On("UpdatePR", mock.Anything, mock.MatchedBy(func(p *models.PullRequest) bool {
			return p.ID == "pr-1" && len(p.Reviewers) == 0
		})).Return(nil)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		require.NoError(t, err)
		mUserRepo.AssertExpectations(t)
		mPRRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		err := svc.DeactivateUsersByTeam(context.Background(), "")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("Error_GetActiveUsers", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(nil, apperrors.ErrInternal)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		assert.Error(t, err)
	})

	t.Run("Error_Deactivate", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		activeUsers := []models.User{
			{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
		}
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
		mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(apperrors.ErrInternal)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		assert.Error(t, err)
	})

	t.Run("Error_GetOpenPRs_Continues", func(t *testing.T) {
		mUserRepo := &mockUserRepoForUserService{}
		mPRRepo := &mockPRRepoForUserService{}
		svc := services.NewUserService(mUserRepo, mPRRepo, log)

		activeUsers := []models.User{
			{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
		}
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
		mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
		mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return(nil, apperrors.ErrInternal)

		err := svc.DeactivateUsersByTeam(context.Background(), "team1")
		require.NoError(t, err)
	})
}
