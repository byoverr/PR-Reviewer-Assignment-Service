package repository

import (
	"context"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUserActive(ctx context.Context, id string, isActive bool) error
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error)
	GetTeamNameByUserID(ctx context.Context, userID string) (string, error)
	DeactivateUsersByTeam(ctx context.Context, teamName string) error
}

type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	UpdatePR(ctx context.Context, pr *models.PullRequest) error
	MergePR(ctx context.Context, id string) error
	GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	ExistsPR(ctx context.Context, id string) (bool, error)
	GetTotalPRs(ctx context.Context) (int, error)
	GetPrsByStatus(ctx context.Context) (int, int, error)
	GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error)
	GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error)
	GetAvgCloseTime(ctx context.Context) (float64, int, error)
	GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error)
	GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error)
}
