package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	log      *slog.Logger
}

var _ TeamServiceInterface = (*TeamService)(nil)

func NewTeamService(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	log *slog.Logger,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		log:      log,
	}
}

// CreateTeam creates a new team.
func (s *TeamService) CreateTeam(ctx context.Context, team *models.Team) (*models.Team, error) {
	if team.Name == "" || len(team.Members) == 0 {
		return nil, apperrors.ErrInvalidInput
	}

	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, apperrors.ErrTeamExists) {
			s.log.InfoContext(ctx, "team already exists, returning error", slog.String("team_name", team.Name))
			return nil, apperrors.ErrTeamExists
		}
		s.log.ErrorContext(ctx, "failed to create team",
			slog.String("team_name", team.Name),
			slog.String("error", err.Error()))
		return nil, apperrors.ErrInternal
	}

	reloaded, err := s.teamRepo.GetTeamByName(ctx, team.Name)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to reload team after create",
			slog.String("team_name", team.Name),
			slog.String("error", err.Error()))
		return nil, apperrors.ErrInternal
	}

	s.log.InfoContext(ctx, "team created successfully",
		slog.String("team_name", team.Name),
		slog.Int("members_count", len(reloaded.Members)))
	return reloaded, nil
}

// AddMemberToTeam upserts a member to an existing team.
func (s *TeamService) AddMemberToTeam(ctx context.Context, teamName string, member models.TeamMember) error {
	if teamName == "" || member.UserID == "" || member.Username == "" {
		return apperrors.ErrInvalidInput
	}

	_, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "team not found for add member", slog.String("team_name", teamName))
			return apperrors.ErrNotFound
		}
		s.log.ErrorContext(ctx, "failed to check team for add member",
			slog.String("team_name", teamName),
			slog.String("error", err.Error()))
		return apperrors.ErrInternal
	}

	user := &models.User{
		ID:       member.UserID,
		Name:     member.Username,
		TeamName: teamName,
		IsActive: member.IsActive,
	}
	if upsertErr := s.userRepo.UpsertUser(ctx, user); upsertErr != nil {
		s.log.ErrorContext(ctx, "failed to add member to team",
			slog.String("team_name", teamName),
			slog.String("user_id", member.UserID),
			slog.String("error", upsertErr.Error()))
		return apperrors.ErrInternal
	}

	s.log.InfoContext(ctx, "member added to team",
		slog.String("team_name", teamName),
		slog.String("user_id", member.UserID))
	return nil
}

// GetTeam retrieves team by name with members.
func (s *TeamService) GetTeam(ctx context.Context, name string) (*models.Team, error) {
	if name == "" {
		return nil, apperrors.ErrInvalidInput
	}

	team, err := s.teamRepo.GetTeamByName(ctx, name)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "team not found", slog.String("team_name", name))
		} else {
			s.log.ErrorContext(ctx, "failed to get team",
				slog.String("team_name", name),
				slog.String("error", err.Error()))
		}
		return nil, err
	}

	s.log.InfoContext(ctx, "team retrieved",
		slog.String("team_name", name),
		slog.Int("members_count", len(team.Members)))
	return team, nil
}
