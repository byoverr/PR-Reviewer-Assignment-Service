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

type mockPRRepo struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepo) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepo) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *mockPRRepo) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepo) MergePR(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPRRepo) GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequestShort), args.Error(1)
}

func (m *mockPRRepo) ExistsPR(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepo) GetTotalPRs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockPRRepo) GetPrsByStatus(ctx context.Context) (int, int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockPRRepo) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepo) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepo) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func (m *mockPRRepo) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepo) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

type mockUserRepo struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepo) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepo) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepo) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func TestPRService_CreatePR(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		pr := &models.PullRequest{ID: "pr-1", Title: "Test", AuthorID: "u1"}

		author := &models.User{ID: "u1", TeamName: "team1"}
		mUserRepo.On("GetUserByID", mock.Anything, "u1").Return(author, nil)
		mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{
			{ID: "u2", IsActive: true},
			{ID: "u3", IsActive: true},
		}, nil)

		mPrRepo.On("CreatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(nil)
		reloaded := &models.PullRequest{ID: "pr-1", Status: "OPEN", Reviewers: []string{"u2", "u3"}}
		mPrRepo.On("GetPRByID", mock.Anything, "pr-1").Return(reloaded, nil)

		result, err := svc.CreatePR(context.Background(), pr)
		require.NoError(t, err)
		assert.Equal(t, "OPEN", result.Status)
		assert.Len(t, result.Reviewers, 2)
		mUserRepo.AssertExpectations(t)
		mPrRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_EmptyID", func(t *testing.T) {
		pr := &models.PullRequest{Title: "Test", AuthorID: "u1"}
		_, err := svc.CreatePR(context.Background(), pr)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyTitle", func(t *testing.T) {
		pr := &models.PullRequest{ID: "pr-1", AuthorID: "u1"}
		_, err := svc.CreatePR(context.Background(), pr)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyAuthorID", func(t *testing.T) {
		pr := &models.PullRequest{ID: "pr-1", Title: "Test"}
		_, err := svc.CreatePR(context.Background(), pr)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("AuthorNotFound", func(t *testing.T) {
		pr := &models.PullRequest{ID: "pr-1", Title: "Test", AuthorID: "u-nonexist"}
		mUserRepo.On("GetUserByID", mock.Anything, "u-nonexist").Return(nil, apperrors.ErrNotFound)

		_, err := svc.CreatePR(context.Background(), pr)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("NoCandidates_OnlyAuthor", func(t *testing.T) {
		mPrRepo2 := &mockPRRepo{}
		mUserRepo2 := &mockUserRepo{}
		svc2 := services.NewPRService(mPrRepo2, mUserRepo2, log)

		pr := &models.PullRequest{ID: "pr-2", Title: "Test", AuthorID: "u1"}
		author := &models.User{ID: "u1", TeamName: "team1"}
		mUserRepo2.On("GetUserByID", mock.Anything, "u1").Return(author, nil)
		mUserRepo2.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{
			{ID: "u1", IsActive: true}, // Only author
		}, nil)

		mPrRepo2.On("CreatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(nil)
		reloaded := &models.PullRequest{ID: "pr-2", Status: "OPEN", Reviewers: []string{}, NeedMoreReviewers: true}
		mPrRepo2.On("GetPRByID", mock.Anything, "pr-2").Return(reloaded, nil)

		result, err := svc2.CreatePR(context.Background(), pr)
		require.NoError(t, err)
		assert.Empty(t, result.Reviewers)
		assert.True(t, result.NeedMoreReviewers)
	})

	t.Run("CreatePR_Failed", func(t *testing.T) {
		mPrRepo3 := &mockPRRepo{}
		mUserRepo3 := &mockUserRepo{}
		svc3 := services.NewPRService(mPrRepo3, mUserRepo3, log)

		pr := &models.PullRequest{ID: "pr-3", Title: "Test", AuthorID: "u1"}
		author := &models.User{ID: "u1", TeamName: "team1"}
		mUserRepo3.On("GetUserByID", mock.Anything, "u1").Return(author, nil)
		mUserRepo3.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{
			{ID: "u2", IsActive: true},
		}, nil)

		mPrRepo3.On("CreatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(apperrors.ErrInternal)

		_, err := svc3.CreatePR(context.Background(), pr)
		assert.Error(t, err)
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
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
			{ID: "u4", IsActive: true},
		}, nil)

		mPrRepo.On("UpdatePR", mock.Anything, mock.AnythingOfType("*models.PullRequest")).Return(nil)

		_, newReviewer, err := svc.ReassignReviewer(context.Background(), "pr-1", "u2")
		require.NoError(t, err)
		assert.NotEqual(t, "u2", newReviewer)
		mPrRepo.AssertExpectations(t)
		mUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_EmptyPRID", func(t *testing.T) {
		_, _, err := svc.ReassignReviewer(context.Background(), "", "u2")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyReviewerID", func(t *testing.T) {
		_, _, err := svc.ReassignReviewer(context.Background(), "pr-1", "")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("PRNotFound", func(t *testing.T) {
		mPrRepo.On("GetPRByID", mock.Anything, "pr-nonexist").Return(nil, apperrors.ErrNotFound)

		_, _, err := svc.ReassignReviewer(context.Background(), "pr-nonexist", "u2")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("PRMerged", func(t *testing.T) {
		mPrRepo4 := &mockPRRepo{}
		mUserRepo4 := &mockUserRepo{}
		svc4 := services.NewPRService(mPrRepo4, mUserRepo4, log)

		pr := &models.PullRequest{ID: "pr-merged", Status: "MERGED"}
		mPrRepo4.On("GetPRByID", mock.Anything, "pr-merged").Return(pr, nil)

		_, _, err := svc4.ReassignReviewer(context.Background(), "pr-merged", "u2")
		assert.ErrorIs(t, err, apperrors.ErrPRMerged)
	})

	t.Run("ReviewerNotAssigned", func(t *testing.T) {
		mPrRepo5 := &mockPRRepo{}
		mUserRepo5 := &mockUserRepo{}
		svc5 := services.NewPRService(mPrRepo5, mUserRepo5, log)

		pr := &models.PullRequest{
			ID:        "pr-notassigned",
			Status:    "OPEN",
			Reviewers: []string{"u2"},
			AuthorID:  "u1",
		}
		mPrRepo5.On("GetPRByID", mock.Anything, "pr-notassigned").Return(pr, nil)
		oldReviewer := &models.User{ID: "u3", TeamName: "team1"}
		mUserRepo5.On("GetUserByID", mock.Anything, "u3").Return(oldReviewer, nil)

		_, _, err := svc5.ReassignReviewer(context.Background(), "pr-notassigned", "u3")
		assert.ErrorIs(t, err, apperrors.ErrNotAssigned)
	})

	t.Run("NoCandidate", func(t *testing.T) {
		mPrRepo6 := &mockPRRepo{}
		mUserRepo6 := &mockUserRepo{}
		svc6 := services.NewPRService(mPrRepo6, mUserRepo6, log)

		pr := &models.PullRequest{
			ID:        "pr-nocandidate",
			Status:    "OPEN",
			Reviewers: []string{"u2"},
			AuthorID:  "u1",
		}
		mPrRepo6.On("GetPRByID", mock.Anything, "pr-nocandidate").Return(pr, nil)
		oldReviewer := &models.User{ID: "u2", TeamName: "team1"}
		mUserRepo6.On("GetUserByID", mock.Anything, "u2").Return(oldReviewer, nil)
		mUserRepo6.On("GetActiveUsersByTeam", mock.Anything, "team1").Return([]models.User{}, nil)

		_, _, err := svc6.ReassignReviewer(context.Background(), "pr-nocandidate", "u2")
		assert.ErrorIs(t, err, apperrors.ErrNoCandidate)
	})
}

func TestPRService_MergePR(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("MergePR", mock.Anything, "pr-1").Return(nil)
		reloaded := &models.PullRequest{ID: "pr-1", Status: "MERGED", Title: "Test"}
		mPrRepo.On("GetPRByID", mock.Anything, "pr-1").Return(reloaded, nil)

		result, err := svc.MergePR(context.Background(), "pr-1")
		require.NoError(t, err)
		assert.Equal(t, "MERGED", result.Status)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		_, err := svc.MergePR(context.Background(), "")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("MergeFailed", func(t *testing.T) {
		mPrRepo7 := &mockPRRepo{}
		mUserRepo7 := &mockUserRepo{}
		svc7 := services.NewPRService(mPrRepo7, mUserRepo7, log)

		mPrRepo7.On("MergePR", mock.Anything, "pr-merge-fail").Return(apperrors.ErrInternal)

		_, err := svc7.MergePR(context.Background(), "pr-merge-fail")
		assert.Error(t, err)
	})

	t.Run("PRNotFoundAfterMerge", func(t *testing.T) {
		mPrRepo8 := &mockPRRepo{}
		mUserRepo8 := &mockUserRepo{}
		svc8 := services.NewPRService(mPrRepo8, mUserRepo8, log)

		mPrRepo8.On("MergePR", mock.Anything, "pr-notfound").Return(nil)
		mPrRepo8.On("GetPRByID", mock.Anything, "pr-notfound").Return(nil, apperrors.ErrNotFound)

		_, err := svc8.MergePR(context.Background(), "pr-notfound")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestPRService_GetTotalPRs(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetTotalPRs", mock.Anything).Return(5, nil)

		total, err := svc.GetTotalPRs(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 5, total)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetTotalPRs", mock.Anything).Return(0, apperrors.ErrInternal)

		_, err := svc9.GetTotalPRs(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetPrsByStatus(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetPrsByStatus", mock.Anything).Return(3, 2, nil)

		open, merged, err := svc.GetPrsByStatus(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 3, open)
		assert.Equal(t, 2, merged)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetPrsByStatus", mock.Anything).Return(0, 0, apperrors.ErrInternal)

		_, _, err := svc9.GetPrsByStatus(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetAssignmentsPerUser(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		assignments := []models.UserAssignment{
			{UserID: "u1", Name: "User1", Count: 5},
			{UserID: "u2", Name: "User2", Count: 3},
		}
		mPrRepo.On("GetAssignmentsPerUser", mock.Anything).Return(assignments, nil)

		result, err := svc.GetAssignmentsPerUser(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetAssignmentsPerUser", mock.Anything).Return(nil, apperrors.ErrInternal)

		_, err := svc9.GetAssignmentsPerUser(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetTopReviewers(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		top := []models.UserAssignment{
			{UserID: "u1", Name: "User1", Count: 10},
		}
		mPrRepo.On("GetTopReviewers", mock.Anything).Return(top, nil)

		result, err := svc.GetTopReviewers(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetTopReviewers", mock.Anything).Return(nil, apperrors.ErrInternal)

		_, err := svc9.GetTopReviewers(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetAvgCloseTime(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		mPrRepo.On("GetAvgCloseTime", mock.Anything).Return(86400.0, 5, nil)

		result, err := svc.GetAvgCloseTime(context.Background())
		require.NoError(t, err)
		assert.InEpsilon(t, 86400.0, result.AverageSeconds, 0.0001)
		assert.Equal(t, 5, result.MergedPRsCount)
		assert.Equal(t, 1, result.Breakdown.Days)
	})

	t.Run("ZeroCount", func(t *testing.T) {
		mPrRepo10 := &mockPRRepo{}
		mUserRepo10 := &mockUserRepo{}
		svc10 := services.NewPRService(mPrRepo10, mUserRepo10, log)

		mPrRepo10.On("GetAvgCloseTime", mock.Anything).Return(0.0, 0, nil)

		result, err := svc10.GetAvgCloseTime(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0.0, result.AverageSeconds)
		assert.Equal(t, 0, result.MergedPRsCount)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetAvgCloseTime", mock.Anything).Return(0.0, 0, apperrors.ErrInternal)

		_, err := svc9.GetAvgCloseTime(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetIdleUsersPerTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		metrics := []models.TeamMetric{
			{TeamName: "team1", Count: 2},
		}
		mPrRepo.On("GetIdleUsersPerTeam", mock.Anything).Return(metrics, nil)

		result, err := svc.GetIdleUsersPerTeam(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetIdleUsersPerTeam", mock.Anything).Return(nil, apperrors.ErrInternal)

		_, err := svc9.GetIdleUsersPerTeam(context.Background())
		assert.Error(t, err)
	})
}

func TestPRService_GetNeedyPRsPerTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mPrRepo := &mockPRRepo{}
	mUserRepo := &mockUserRepo{}

	svc := services.NewPRService(mPrRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		metrics := []models.TeamMetric{
			{TeamName: "team1", Count: 3},
		}
		mPrRepo.On("GetNeedyPRsPerTeam", mock.Anything).Return(metrics, nil)

		result, err := svc.GetNeedyPRsPerTeam(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("Error", func(t *testing.T) {
		mPrRepo9 := &mockPRRepo{}
		mUserRepo9 := &mockUserRepo{}
		svc9 := services.NewPRService(mPrRepo9, mUserRepo9, log)

		mPrRepo9.On("GetNeedyPRsPerTeam", mock.Anything).Return(nil, apperrors.ErrInternal)

		_, err := svc9.GetNeedyPRsPerTeam(context.Background())
		assert.Error(t, err)
	})
}
