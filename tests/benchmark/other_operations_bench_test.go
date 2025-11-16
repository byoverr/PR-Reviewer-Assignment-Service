package benchmark_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"
	"github.com/stretchr/testify/mock"
)

type mockUserRepoForPRBench struct {
	mock.Mock
	repository.UserRepository
}

func (m *mockUserRepoForPRBench) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepoForPRBench) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

type mockPRRepoForOpsBench struct {
	mock.Mock
	repository.PRRepository
}

func (m *mockPRRepoForOpsBench) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForOpsBench) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *mockPRRepoForOpsBench) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *mockPRRepoForOpsBench) MergePR(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPRRepoForOpsBench) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UserAssignment), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetPrsByStatus(ctx context.Context) (int, int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForOpsBench) GetTotalPRs(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMetric), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Int(1), args.Error(2)
}

func (m *mockPRRepoForOpsBench) GetOpenPRsWithReviewersFromTeam(
	ctx context.Context, teamName string,
) ([]models.PullRequest, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequest), args.Error(1)
}

func (m *mockPRRepoForOpsBench) ExistsPR(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepoForOpsBench) GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.PullRequestShort), args.Error(1)
}

// BenchmarkCreatePR benchmarks PR creation with auto-assignment.
func BenchmarkCreatePR(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	author := &models.User{
		ID:       "author1",
		Name:     "Author",
		TeamName: "team1",
		IsActive: true,
	}

	activeUsers := []models.User{
		{ID: "u1", Name: "User1", TeamName: "team1", IsActive: true},
		{ID: "u2", Name: "User2", TeamName: "team1", IsActive: true},
		{ID: "u3", Name: "User3", TeamName: "team1", IsActive: true},
	}

	createdPR := &models.PullRequest{
		ID:        "pr-1",
		Title:     "Test PR",
		AuthorID:  "author1",
		Status:    "OPEN",
		Reviewers: []string{"u1", "u2"},
	}
	now := time.Now()
	createdPR.CreatedAt = &now

	mUserRepo.On("GetUserByID", mock.Anything, "author1").Return(author, nil)
	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
	mPRRepo.On("ExistsPR", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mPRRepo.On("CreatePR", mock.Anything, mock.Anything).Return(nil)
	mPRRepo.On("GetPRByID", mock.Anything, mock.AnythingOfType("string")).Return(createdPR, nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for i := range b.N {
		pr := &models.PullRequest{
			ID:       fmt.Sprintf("pr-%d", i),
			Title:    fmt.Sprintf("PR %d", i),
			AuthorID: "author1",
		}
		_, _ = svc.CreatePR(context.Background(), pr)
	}
}

// BenchmarkReassignReviewer benchmarks reviewer reassignment.
func BenchmarkReassignReviewer(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	pr := &models.PullRequest{
		ID:        "pr-1",
		Title:     "Test PR",
		AuthorID:  "author1",
		Status:    "OPEN",
		Reviewers: []string{"u1", "u2"},
	}

	oldReviewer := &models.User{
		ID:       "u1",
		Name:     "User1",
		TeamName: "team1",
		IsActive: true,
	}

	activeUsers := []models.User{
		{ID: "u3", Name: "User3", TeamName: "team1", IsActive: true},
		{ID: "u4", Name: "User4", TeamName: "team1", IsActive: true},
	}

	mPRRepo.On("GetPRByID", mock.Anything, "pr-1").Return(pr, nil)
	mUserRepo.On("GetUserByID", mock.Anything, "u1").Return(oldReviewer, nil)
	mUserRepo.On("GetActiveUsersByTeam", mock.Anything, "team1").Return(activeUsers, nil)
	mPRRepo.On("UpdatePR", mock.Anything, mock.Anything).Return(nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_, _, _ = svc.ReassignReviewer(context.Background(), "pr-1", "u1")
	}
}

// BenchmarkGetTopReviewers benchmarks top reviewers retrieval.
func BenchmarkGetTopReviewers(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	topReviewers := []models.UserAssignment{
		{UserID: "u1", Name: "User1", Count: 10},
		{UserID: "u2", Name: "User2", Count: 8},
		{UserID: "u3", Name: "User3", Count: 6},
		{UserID: "u4", Name: "User4", Count: 4},
		{UserID: "u5", Name: "User5", Count: 2},
	}

	mPRRepo.On("GetTopReviewers", mock.Anything).Return(topReviewers, nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.GetTopReviewers(context.Background())
	}
}

// BenchmarkGetAssignmentsPerUser benchmarks assignments per user retrieval.
func BenchmarkGetAssignmentsPerUser(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	assignments := make([]models.UserAssignment, 50)
	for i := range 50 {
		assignments[i] = models.UserAssignment{
			UserID: fmt.Sprintf("u%d", i),
			Name:   fmt.Sprintf("User%d", i),
			Count:  10 - i/5,
		}
	}

	mPRRepo.On("GetAssignmentsPerUser", mock.Anything).Return(assignments, nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_, _ = svc.GetAssignmentsPerUser(context.Background())
	}
}

// BenchmarkMergePR benchmarks PR merging.
func BenchmarkMergePR(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	pr := &models.PullRequest{
		ID:        "pr-1",
		Title:     "Test PR",
		AuthorID:  "author1",
		Status:    "OPEN",
		Reviewers: []string{"u1", "u2"},
	}

	mergedPR := *pr
	mergedPR.Status = "MERGED"
	now := time.Now()
	mergedPR.MergedAt = &now

	mPRRepo.On("MergePR", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	mPRRepo.On("GetPRByID", mock.Anything, mock.AnythingOfType("string")).Return(&mergedPR, nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for i := range b.N {
		_, _ = svc.MergePR(context.Background(), fmt.Sprintf("pr-%d", i))
	}
}

// BenchmarkGetPrsByStatus benchmarks PR status retrieval.
func BenchmarkGetPrsByStatus(b *testing.B) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	mUserRepo := &mockUserRepoForPRBench{}
	mPRRepo := &mockPRRepoForOpsBench{}

	mPRRepo.On("GetPrsByStatus", mock.Anything).Return(100, 50, nil)

	svc := services.NewPRService(mPRRepo, mUserRepo, log)

	b.ResetTimer()
	for b.Loop() {
		_, _, _ = svc.GetPrsByStatus(context.Background())
	}
}
