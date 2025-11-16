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

type mockTeamRepo struct {
	mock.Mock
	repository.TeamRepository
}

func (m *mockTeamRepo) CreateTeam(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamRepo) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

type mockUserRepoForTeamService struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoForTeamService) UpsertUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoForTeamService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForTeamService) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	args := m.Called(ctx, id, isActive)
	return args.Error(0)
}

func (m *mockUserRepoForTeamService) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoForTeamService) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *mockUserRepoForTeamService) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func TestTeamService_CreateTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepo{}
	mUserRepo := &mockUserRepoForTeamService{}

	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
				{UserID: "u2", Username: "User2", IsActive: true},
			},
		}
		mTeamRepo.On("CreateTeam", mock.Anything, team).Return(nil)
		reloaded := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
				{UserID: "u2", Username: "User2", IsActive: true},
			},
		}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").Return(reloaded, nil)

		result, err := svc.CreateTeam(context.Background(), team)
		require.NoError(t, err)
		assert.Equal(t, "team1", result.Name)
		assert.Len(t, result.Members, 2)
		mTeamRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_EmptyName", func(t *testing.T) {
		team := &models.Team{
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1"},
			},
		}
		_, err := svc.CreateTeam(context.Background(), team)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyMembers", func(t *testing.T) {
		team := &models.Team{
			Name:    "team1",
			Members: []models.TeamMember{},
		}
		_, err := svc.CreateTeam(context.Background(), team)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("TeamExists", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1"},
			},
		}
		mTeamRepo.On("CreateTeam", mock.Anything, team).Return(apperrors.ErrTeamExists)

		_, err := svc.CreateTeam(context.Background(), team)
		assert.ErrorIs(t, err, apperrors.ErrTeamExists)
	})

	t.Run("ReloadFailed", func(t *testing.T) {
		mTeamRepo2 := &mockTeamRepo{}
		mUserRepo2 := &mockUserRepoForTeamService{}
		svc2 := services.NewTeamService(mTeamRepo2, mUserRepo2, log)

		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1"},
			},
		}
		mTeamRepo2.On("CreateTeam", mock.Anything, team).Return(nil)
		mTeamRepo2.On("GetTeamByName", mock.Anything, "team1").Return(nil, apperrors.ErrInternal)

		_, err := svc2.CreateTeam(context.Background(), team)
		assert.ErrorIs(t, err, apperrors.ErrInternal)
	})
}

func TestTeamService_AddMemberToTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepo{}
	mUserRepo := &mockUserRepoForTeamService{}

	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{Name: "team1"}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").Return(team, nil)
		member := models.TeamMember{UserID: "u1", Username: "User1", IsActive: true}
		mUserRepo.On("UpsertUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

		err := svc.AddMemberToTeam(context.Background(), "team1", member)
		require.NoError(t, err)
		mTeamRepo.AssertExpectations(t)
		mUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput_EmptyTeamName", func(t *testing.T) {
		member := models.TeamMember{UserID: "u1", Username: "User1"}
		err := svc.AddMemberToTeam(context.Background(), "", member)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyUserID", func(t *testing.T) {
		member := models.TeamMember{Username: "User1"}
		err := svc.AddMemberToTeam(context.Background(), "team1", member)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("InvalidInput_EmptyUsername", func(t *testing.T) {
		member := models.TeamMember{UserID: "u1"}
		err := svc.AddMemberToTeam(context.Background(), "team1", member)
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("TeamNotFound", func(t *testing.T) {
		mTeamRepo3 := &mockTeamRepo{}
		mUserRepo3 := &mockUserRepoForTeamService{}
		svc3 := services.NewTeamService(mTeamRepo3, mUserRepo3, log)

		mTeamRepo3.On("GetTeamByName", mock.Anything, "team-nonexist").Return(nil, apperrors.ErrNotFound)
		member := models.TeamMember{UserID: "u1", Username: "User1"}

		err := svc3.AddMemberToTeam(context.Background(), "team-nonexist", member)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("UpsertUserFailed", func(t *testing.T) {
		mTeamRepo4 := &mockTeamRepo{}
		mUserRepo4 := &mockUserRepoForTeamService{}
		svc4 := services.NewTeamService(mTeamRepo4, mUserRepo4, log)

		team := &models.Team{Name: "team1"}
		mTeamRepo4.On("GetTeamByName", mock.Anything, "team1").Return(team, nil)
		member := models.TeamMember{UserID: "u1", Username: "User1"}
		mUserRepo4.On("UpsertUser", mock.Anything, mock.AnythingOfType("*models.User")).Return(apperrors.ErrInternal)

		err := svc4.AddMemberToTeam(context.Background(), "team1", member)
		assert.ErrorIs(t, err, apperrors.ErrInternal)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mTeamRepo := &mockTeamRepo{}
	mUserRepo := &mockUserRepoForTeamService{}

	svc := services.NewTeamService(mTeamRepo, mUserRepo, log)

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
			},
		}
		mTeamRepo.On("GetTeamByName", mock.Anything, "team1").Return(team, nil)

		result, err := svc.GetTeam(context.Background(), "team1")
		require.NoError(t, err)
		assert.Equal(t, "team1", result.Name)
		assert.Len(t, result.Members, 1)
		mTeamRepo.AssertExpectations(t)
	})

	t.Run("InvalidInput", func(t *testing.T) {
		_, err := svc.GetTeam(context.Background(), "")
		assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
	})

	t.Run("TeamNotFound", func(t *testing.T) {
		mTeamRepo5 := &mockTeamRepo{}
		mUserRepo5 := &mockUserRepoForTeamService{}
		svc5 := services.NewTeamService(mTeamRepo5, mUserRepo5, log)

		mTeamRepo5.On("GetTeamByName", mock.Anything, "team-nonexist").Return(nil, apperrors.ErrNotFound)

		_, err := svc5.GetTeam(context.Background(), "team-nonexist")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("Error", func(t *testing.T) {
		mTeamRepo6 := &mockTeamRepo{}
		mUserRepo6 := &mockUserRepoForTeamService{}
		svc6 := services.NewTeamService(mTeamRepo6, mUserRepo6, log)

		mTeamRepo6.On("GetTeamByName", mock.Anything, "team1").Return(nil, apperrors.ErrInternal)

		_, err := svc6.GetTeam(context.Background(), "team1")
		assert.Error(t, err)
	})
}
