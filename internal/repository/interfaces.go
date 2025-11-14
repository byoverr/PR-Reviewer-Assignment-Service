package repository

import (
	"context"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

// TeamRepository интерфейс для операций с командами
type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
}

// UserRepository интерфейс для операций с пользователями
type UserRepository interface {
	UpsertUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	UpdateUserActive(ctx context.Context, id string, isActive bool) error
	GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error)
	GetTeamNameByUserID(ctx context.Context, userID string) (string, error)
}

// PRRepository интерфейс для операций с PR
type PRRepository interface {
	CreatePR(ctx context.Context, pr *models.PullRequest) error
	GetPRByID(ctx context.Context, id string) (*models.PullRequest, error)
	UpdatePR(ctx context.Context, pr *models.PullRequest) error
	MergePR(ctx context.Context, id string) error
	GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	ExistsPR(ctx context.Context, id string) (bool, error)
}
