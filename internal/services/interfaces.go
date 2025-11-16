package services

import (
	"context"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

type TeamServiceInterface interface {
	CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error)
	GetTeam(ctx context.Context, name string) (*models.Team, error)
}

type UserServiceInterface interface {
	SetUserActive(ctx context.Context, id string, isActive bool) (*models.User, error)
	GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	DeactivateUsersByTeam(ctx context.Context, teamName string) error
}

type PRServiceInterface interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	GetTotalPRs(ctx context.Context) (int, error)
	GetPrsByStatus(ctx context.Context) (int, int, error)
	GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error)
	GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error)
	GetAvgCloseTime(ctx context.Context) (models.AvgCloseTimeDetail, error)
	GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error)
	GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error)
}
