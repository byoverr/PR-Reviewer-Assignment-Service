package benchmark_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/stretchr/testify/mock"
)

type mockUserRepoBench struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoBench) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepoBench) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

type mockPRRepoBench struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepoBench) GetOpenPRsWithReviewersFromTeam(
	ctx context.Context, teamName string,
) ([]models.PullRequest, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *mockPRRepoBench) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

// BenchmarkDeactivateUsersByTeam_NoPRs benchmarks deactivation with no PRs to reassign.
func BenchmarkDeactivateUsersByTeam_NoPRs(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoBench{}
	mPRRepo := &mockPRRepoBench{}

	// Setup: 10 active users in team
	activeUsers := make([]models.User, 10)
	for i := range 10 {
		activeUsers[i] = models.User{
			ID:       fmt.Sprintf("u%d", i),
			Name:     fmt.Sprintf("User%d", i),
			TeamName: "team1",
			IsActive: true,
		}
	}

	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
	mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
	mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return([]models.PullRequest{}, nil)

	svc := services.NewUserService(mUserRepo, mPRRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_ = svc.DeactivateUsersByTeam(context.Background(), "team1")
	}
}

// BenchmarkDeactivateUsersByTeam_WithPRs benchmarks deactivation with PRs to reassign.
func BenchmarkDeactivateUsersByTeam_WithPRs(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoBench{}
	mPRRepo := &mockPRRepoBench{}

	// Setup: 20 active users in team
	activeUsers := make([]models.User, 20)
	for i := range 20 {
		activeUsers[i] = models.User{
			ID:       fmt.Sprintf("u%d", i),
			Name:     fmt.Sprintf("User%d", i),
			TeamName: "team1",
			IsActive: true,
		}
	}

	// Setup: 5 active users remain after deactivation (for replacement)
	activeUsersAfter := make([]models.User, 5)
	for i := range 5 {
		activeUsersAfter[i] = models.User{
			ID:       fmt.Sprintf("u%d", i+20),
			Name:     fmt.Sprintf("User%d", i+20),
			TeamName: "team1",
			IsActive: true,
		}
	}

	// Setup: 10 open PRs with reviewers from the team
	prs := make([]models.PullRequest, 10)
	for i := range 10 {
		prs[i] = models.PullRequest{
			ID:        fmt.Sprintf("pr-%d", i),
			Title:     fmt.Sprintf("Test PR %d", i),
			AuthorID:  "author1",
			Status:    "OPEN",
			Reviewers: []string{"u0", "u1"}, // Will be deactivated
		}
	}

	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
	mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
	mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return(prs, nil)
	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsersAfter, nil)
	mPRRepo.On("UpdatePR", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewUserService(mUserRepo, mPRRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_ = svc.DeactivateUsersByTeam(context.Background(), "team1")
	}
}

// BenchmarkDeactivateUsersByTeam_LargeScale benchmarks with large number of users and PRs.
func BenchmarkDeactivateUsersByTeam_LargeScale(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoBench{}
	mPRRepo := &mockPRRepoBench{}

	// Setup: 100 active users in team
	activeUsers := make([]models.User, 100)
	for i := range 100 {
		activeUsers[i] = models.User{
			ID:       fmt.Sprintf("u%02d", i),
			Name:     fmt.Sprintf("User%02d", i),
			TeamName: "team1",
			IsActive: true,
		}
	}

	// Setup: 20 active users remain after deactivation
	activeUsersAfter := make([]models.User, 20)
	for i := range 20 {
		activeUsersAfter[i] = models.User{
			ID:       fmt.Sprintf("u%02d", i+100),
			Name:     fmt.Sprintf("User%02d", i+100),
			TeamName: "team1",
			IsActive: true,
		}
	}

	// Setup: 50 open PRs with reviewers from the team
	prs := make([]models.PullRequest, 50)
	for i := range 50 {
		prs[i] = models.PullRequest{
			ID:        fmt.Sprintf("pr-%02d", i),
			Title:     fmt.Sprintf("Test PR %02d", i),
			AuthorID:  "author1",
			Status:    "OPEN",
			Reviewers: []string{"u00", "u01"}, // Will be deactivated
		}
	}

	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
	mUserRepo.On("DeactivateUsersByTeam", mock.Anything, "team1").Return(nil)
	mPRRepo.On("GetOpenPRsWithReviewersFromTeam", mock.Anything, "team1").Return(prs, nil)
	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsersAfter, nil)
	mPRRepo.On("UpdatePR", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewUserService(mUserRepo, mPRRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_ = svc.DeactivateUsersByTeam(context.Background(), "team1")
	}
}
