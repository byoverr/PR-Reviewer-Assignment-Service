package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
	log      *slog.Logger
}

var _ UserServiceInterface = (*UserService)(nil)

func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository, log *slog.Logger) *UserService {
	return &UserService{
		userRepo: userRepo,
		prRepo:   prRepo,
		log:      log,
	}
}

// SetUserActive updates is_active flag.
func (s *UserService) SetUserActive(ctx context.Context, id string, isActive bool) (*models.User, error) {
	if id == "" {
		return nil, apperrors.ErrInvalidInput
	}

	err := s.userRepo.UpdateUserActive(ctx, id, isActive)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "user not found for active update", slog.String("user_id", id))
		} else {
			s.log.ErrorContext(ctx, "failed to update user active",
				slog.String("user_id", id),
				slog.Bool("is_active", isActive),
				slog.String("error", err.Error()))
		}
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to reload user after update",
			slog.String("user_id", id),
			slog.String("error", err.Error()))
		return nil, apperrors.ErrInternal
	}

	s.log.InfoContext(ctx, "user active updated",
		slog.String("user_id", id),
		slog.Bool("is_active", isActive))
	return user, nil
}

// GetPRsForUser returns PRs assigned to user as reviewer.
func (s *UserService) GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	if userID == "" {
		return nil, apperrors.ErrInvalidInput
	}

	prs, err := s.prRepo.GetPRsForUser(ctx, userID)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get PRs for user",
			slog.String("user_id", userID),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.log.InfoContext(ctx, "PRs retrieved for user",
		slog.String("user_id", userID),
		slog.Int("count", len(prs)))
	return prs, nil
}

// DeactivateUsersByTeam mass deactivates + reassign PRs if open.
func (s *UserService) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
	if teamName == "" {
		return apperrors.ErrInvalidInput
	}

	deactivateErr := s.userRepo.DeactivateUsersByTeam(ctx, teamName)
	if deactivateErr != nil {
		s.log.ErrorContext(ctx, "failed to deactivate team users",
			slog.String("team_name", teamName),
			slog.String("error", deactivateErr.Error()))
		return deactivateErr
	}

	s.log.InfoContext(ctx, "team users deactivated", slog.String("team_name", teamName))

	return nil
}
