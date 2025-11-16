package services

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"

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

	activeUsersBefore, err := s.userRepo.GetActiveUsersByTeam(ctx, teamName)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get active users before deactivation",
			slog.String("team_name", teamName),
			slog.String("error", err.Error()))
		return apperrors.Wrap(err, "failed to get active users")
	}

	deactivatedUserIDs := make(map[string]bool, len(activeUsersBefore))
	for _, u := range activeUsersBefore {
		deactivatedUserIDs[u.ID] = true
	}

	// Deactivate users
	deactivateErr := s.userRepo.DeactivateUsersByTeam(ctx, teamName)
	if deactivateErr != nil {
		s.log.ErrorContext(ctx, "failed to deactivate team users",
			slog.String("team_name", teamName),
			slog.String("error", deactivateErr.Error()))
		return deactivateErr
	}

	openPRs, err := s.prRepo.GetOpenPRsWithReviewersFromTeam(ctx, teamName)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get open PRs for reassignment",
			slog.String("team_name", teamName),
			slog.String("error", err.Error()))
		return nil
	}

	reassignedCount := 0
	for _, pr := range openPRs {
		reassigned, reassignErr := s.reassignDeactivatedReviewers(ctx, &pr, deactivatedUserIDs, teamName)
		if reassignErr != nil {
			s.log.WarnContext(ctx, "failed to reassign reviewers for PR",
				slog.String("pr_id", pr.ID),
				slog.String("error", reassignErr.Error()))
			continue
		}
		if reassigned {
			reassignedCount++
		}
	}

	s.log.InfoContext(ctx, "team users deactivated and PRs reassigned",
		slog.String("team_name", teamName),
		slog.Int("deactivated_users", len(deactivatedUserIDs)),
		slog.Int("prs_processed", len(openPRs)),
		slog.Int("prs_reassigned", reassignedCount))

	return nil
}

// reassignDeactivatedReviewers reassigns deactivated reviewers in a PR to active team members.
func (s *UserService) reassignDeactivatedReviewers(
	ctx context.Context,
	pr *models.PullRequest,
	deactivatedUserIDs map[string]bool,
	teamName string,
) (bool, error) {
	deactivatedReviewers := s.findDeactivatedReviewers(pr.Reviewers, deactivatedUserIDs)
	if len(deactivatedReviewers) == 0 {
		return false, nil
	}

	activeUsers, err := s.userRepo.GetActiveUsersByTeam(ctx, teamName)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to get active users for reassignment")
	}

	candidates := s.buildCandidateList(pr, activeUsers)
	if len(candidates) == 0 {
		return s.removeDeactivatedReviewers(ctx, pr, deactivatedUserIDs)
	}

	updated := s.replaceReviewers(pr, deactivatedReviewers, candidates)
	if !updated {
		return false, nil
	}

	const minReviewers = 2
	pr.NeedMoreReviewers = len(pr.Reviewers) < minReviewers

	if updateErr := s.prRepo.UpdatePR(ctx, pr); updateErr != nil {
		return false, apperrors.Wrap(updateErr, "failed to update PR after reassignment")
	}

	return true, nil
}

// findDeactivatedReviewers finds reviewers that are in the deactivated list.
func (s *UserService) findDeactivatedReviewers(reviewers []string, deactivatedUserIDs map[string]bool) []string {
	deactivatedReviewers := []string{}
	for _, reviewerID := range reviewers {
		if deactivatedUserIDs[reviewerID] {
			deactivatedReviewers = append(deactivatedReviewers, reviewerID)
		}
	}
	return deactivatedReviewers
}

// buildCandidateList builds a list of candidate reviewers from active team members.
func (s *UserService) buildCandidateList(pr *models.PullRequest, activeUsers []models.User) []string {
	exclude := map[string]bool{pr.AuthorID: true}
	for _, r := range pr.Reviewers {
		exclude[r] = true
	}

	candidates := []string{}
	for _, u := range activeUsers {
		if !exclude[u.ID] {
			candidates = append(candidates, u.ID)
		}
	}
	return candidates
}

// removeDeactivatedReviewers removes deactivated reviewers from PR when no replacements are available.
func (s *UserService) removeDeactivatedReviewers(
	ctx context.Context,
	pr *models.PullRequest,
	deactivatedUserIDs map[string]bool,
) (bool, error) {
	newReviewers := []string{}
	for _, r := range pr.Reviewers {
		if !deactivatedUserIDs[r] {
			newReviewers = append(newReviewers, r)
		}
	}
	pr.Reviewers = newReviewers
	const minReviewers = 2
	pr.NeedMoreReviewers = len(pr.Reviewers) < minReviewers

	if updateErr := s.prRepo.UpdatePR(ctx, pr); updateErr != nil {
		return false, apperrors.Wrap(updateErr, "failed to update PR after removing deactivated reviewers")
	}
	return true, nil
}

// replaceReviewers replaces deactivated reviewers with candidates.
func (s *UserService) replaceReviewers(
	pr *models.PullRequest,
	deactivatedReviewers []string,
	candidates []string,
) bool {
	exclude := map[string]bool{pr.AuthorID: true}
	for _, r := range pr.Reviewers {
		exclude[r] = true
	}

	updated := false
	for _, oldReviewerID := range deactivatedReviewers {
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		newReviewerID := s.findAvailableCandidate(candidates, exclude)
		if newReviewerID == "" {
			s.removeReviewer(pr, oldReviewerID)
		} else {
			s.replaceReviewer(pr, oldReviewerID, newReviewerID)
			exclude[newReviewerID] = true
		}
		updated = true
	}

	return updated
}

// findAvailableCandidate finds the first available candidate that's not excluded.
func (s *UserService) findAvailableCandidate(candidates []string, exclude map[string]bool) string {
	for _, candidateID := range candidates {
		if !exclude[candidateID] {
			return candidateID
		}
	}
	return ""
}

// removeReviewer removes a reviewer from PR.
func (s *UserService) removeReviewer(pr *models.PullRequest, reviewerID string) {
	newReviewers := []string{}
	for _, r := range pr.Reviewers {
		if r != reviewerID {
			newReviewers = append(newReviewers, r)
		}
	}
	pr.Reviewers = newReviewers
}

// replaceReviewer replaces old reviewer with new one in PR.
func (s *UserService) replaceReviewer(pr *models.PullRequest, oldReviewerID, newReviewerID string) {
	for i, r := range pr.Reviewers {
		if r == oldReviewerID {
			pr.Reviewers[i] = newReviewerID
			break
		}
	}
}
