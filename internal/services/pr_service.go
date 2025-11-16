package services

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
)

type PRService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	log      *slog.Logger
}

var _ PRServiceInterface = (*PRService)(nil)

func NewPRService(prRepo repository.PRRepository, userRepo repository.UserRepository, log *slog.Logger) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		log:      log,
	}
}

// CreatePR creates PR and auto-assigns up to 2 active reviewers from author's team (exclude author).
func (s *PRService) CreatePR(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	if pr.ID == "" || pr.Title == "" || pr.AuthorID == "" {
		return nil, apperrors.ErrInvalidInput
	}

	pr.Status = "OPEN"

	author, err := s.userRepo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "author not found for PR create", slog.String("author_id", pr.AuthorID))
		} else {
			s.log.ErrorContext(ctx, "failed to get author",
				slog.String("author_id", pr.AuthorID),
				slog.String("error", err.Error()))
		}
		return nil, apperrors.Wrap(err, "author validation failed")
	}

	activeUsers, err := s.userRepo.GetActiveUsersByTeam(ctx, author.TeamName)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get active team users",
			slog.String("team_name", author.TeamName),
			slog.String("error", err.Error()))
		return nil, apperrors.Wrap(err, "team users fetch failed")
	}

	candidates := make([]string, 0, len(activeUsers))
	for _, u := range activeUsers {
		if u.ID != pr.AuthorID {
			candidates = append(candidates, u.ID)
		}
	}

	// Shuffle candidates
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	const maxReviewers = 2

	numReviewers := min(len(candidates), maxReviewers)
	pr.Reviewers = candidates[:numReviewers]
	pr.NeedMoreReviewers = numReviewers < maxReviewers
	now := time.Now()
	pr.CreatedAt = &now

	if createErr := s.prRepo.CreatePR(ctx, pr); createErr != nil {
		s.log.ErrorContext(ctx, "failed to create PR",
			slog.String("pr_id", pr.ID),
			slog.String("error", createErr.Error()))
		return nil, createErr
	}

	reloaded, err := s.prRepo.GetPRByID(ctx, pr.ID)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to reload PR after create",
			slog.String("pr_id", pr.ID),
			slog.String("error", err.Error()))
		return nil, apperrors.ErrInternal
	}

	s.log.InfoContext(ctx, "PR created with auto-assign",
		slog.String("pr_id", pr.ID),
		slog.Int("reviewers_count", len(reloaded.Reviewers)),
		slog.Bool("need_more", reloaded.NeedMoreReviewers))
	return reloaded, nil
}

// ReassignReviewer replaces old_reviewer_id with random active from old's team (exclude current/author).
func (s *PRService) ReassignReviewer(
	ctx context.Context,
	prID, oldReviewerID string,
) (*models.PullRequest, string, error) {
	if prID == "" || oldReviewerID == "" {
		return nil, "", apperrors.ErrInvalidInput
	}

	pr, err := s.getPRForReassign(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	newReviewer, err := s.selectNewReviewer(ctx, pr, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	s.replaceReviewerInPR(pr, oldReviewerID, newReviewer)

	if updateErr := s.prRepo.UpdatePR(ctx, pr); updateErr != nil {
		s.log.ErrorContext(ctx, "failed to update PR for reassign",
			slog.String("pr_id", prID),
			slog.String("error", updateErr.Error()))
		return nil, "", updateErr
	}

	s.log.InfoContext(ctx, "reviewer reassigned",
		slog.String("pr_id", prID),
		slog.String("old", oldReviewerID),
		slog.String("new", newReviewer))
	return pr, newReviewer, nil
}

// getPRForReassign проверяет существование PR, статус и наличие старого ревьюера.
func (s *PRService) getPRForReassign(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "PR not found for reassign", slog.String("pr_id", prID))
		}
		return nil, err
	}
	if pr.Status == "MERGED" {
		return nil, apperrors.ErrPRMerged
	}
	return pr, nil
}

// selectNewReviewer selects a new reviewer from the active team members.
func (s *PRService) selectNewReviewer(
	ctx context.Context,
	pr *models.PullRequest,
	oldReviewerID string,
) (string, error) {
	if !slices.Contains(pr.Reviewers, oldReviewerID) {
		return "", apperrors.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get old reviewer",
			slog.String("reviewer_id", oldReviewerID),
			slog.String("error", err.Error()))
		return "", apperrors.Wrap(err, "old reviewer fetch failed")
	}

	activeInTeam, err := s.userRepo.GetActiveUsersByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get active team users for reassign",
			slog.String("team_name", oldReviewer.TeamName),
			slog.String("error", err.Error()))
		return "", apperrors.Wrap(err, "team users fetch failed")
	}

	exclude := map[string]bool{pr.AuthorID: true}
	for _, r := range pr.Reviewers {
		exclude[r] = true
	}

	candidates := []string{}
	for _, u := range activeInTeam {
		if !exclude[u.ID] {
			candidates = append(candidates, u.ID)
		}
	}

	if len(candidates) == 0 {
		return "", apperrors.ErrNoCandidate
	}

	//nolint:gosec // for this app is allowed to use rand/v2
	newReviewer := candidates[rand.IntN(len(candidates))]
	return newReviewer, nil
}

// replaceReviewerInPR replaces the old reviewer with the new one and updates the NeedMoreReviewers flag.
func (s *PRService) replaceReviewerInPR(pr *models.PullRequest, oldReviewerID, newReviewer string) {
	const minReviewers = 2

	for i, r := range pr.Reviewers {
		if r == oldReviewerID {
			pr.Reviewers[i] = newReviewer
			break
		}
	}
	pr.NeedMoreReviewers = len(pr.Reviewers) < minReviewers
}

// MergePR sets status to MERGED.
func (s *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	if prID == "" {
		return nil, apperrors.ErrInvalidInput
	}

	err := s.prRepo.MergePR(ctx, prID)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to merge PR",
			slog.String("pr_id", prID),
			slog.String("error", err.Error()))
		return nil, err
	}

	pr, err := s.prRepo.GetPRByID(ctx, prID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			s.log.WarnContext(ctx, "PR not found after merge", slog.String("pr_id", prID))
		}
		return nil, err
	}

	s.log.InfoContext(ctx, "PR merged", slog.String("pr_id", prID), slog.String("title", pr.Title))
	return pr, nil
}

// GetTotalPRs returns the total count of all pull requests.
func (s *PRService) GetTotalPRs(ctx context.Context) (int, error) {
	total, err := s.prRepo.GetTotalPRs(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get total PRs", slog.String("error", err.Error()))
		return 0, err
	}
	s.log.InfoContext(ctx, "total PRs fetched", slog.Int("total", total))
	return total, nil
}

// GetPrsByStatus returns the count of open and merged pull requests.
func (s *PRService) GetPrsByStatus(ctx context.Context) (int, int, error) {
	open, merged, err := s.prRepo.GetPrsByStatus(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get PRs by status", slog.String("error", err.Error()))
		return 0, 0, err
	}
	s.log.InfoContext(ctx, "PRs by status fetched", slog.Int("open", open), slog.Int("merged", merged))
	return open, merged, nil
}

// GetAssignmentsPerUser returns the number of PR assignments per active user, ordered by count descending.
func (s *PRService) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	assignments, err := s.prRepo.GetAssignmentsPerUser(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get assignments per user", slog.String("error", err.Error()))
		return nil, err
	}
	s.log.InfoContext(ctx, "assignments per user fetched", slog.Int("total_users", len(assignments)))
	return assignments, nil
}

// GetTopReviewers returns the top 5 reviewers by assignment count.
func (s *PRService) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	top, err := s.prRepo.GetTopReviewers(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get top reviewers", slog.String("error", err.Error()))
		return nil, err
	}
	s.log.InfoContext(ctx, "top reviewers fetched", slog.Int("top_count", len(top)))
	return top, nil
}

// GetAvgCloseTime returns the average time to close merged PRs with breakdown by days, hours, minutes, and seconds.
func (s *PRService) GetAvgCloseTime(ctx context.Context) (models.AvgCloseTimeDetail, error) {
	const (
		secondsPerDay    = 86400
		secondsPerHour   = 3600
		secondsPerMinute = 60
	)

	avgSeconds, count, err := s.prRepo.GetAvgCloseTime(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get avg close time", slog.String("error", err.Error()))
		return models.AvgCloseTimeDetail{}, err
	}

	detail := models.AvgCloseTimeDetail{
		AverageSeconds: avgSeconds,
		MergedPRsCount: count,
	}
	if count > 0 {
		detail.Breakdown.Days = int(avgSeconds / secondsPerDay)
		remainder := avgSeconds - float64(detail.Breakdown.Days*secondsPerDay)
		detail.Breakdown.Hours = int(remainder / secondsPerHour)
		remainder -= float64(detail.Breakdown.Hours * secondsPerHour)
		detail.Breakdown.Minutes = int(remainder / secondsPerMinute)
		detail.Breakdown.Seconds = int(remainder - float64(detail.Breakdown.Minutes*secondsPerMinute))
	}

	s.log.InfoContext(ctx, "avg close time fetched",
		slog.Float64("seconds", detail.AverageSeconds),
		slog.Int("merged_count", count))
	return detail, nil
}

// GetIdleUsersPerTeam returns the count of active users with zero PR assignments, grouped by team.
func (s *PRService) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	metrics, err := s.prRepo.GetIdleUsersPerTeam(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get idle users per team", slog.String("error", err.Error()))
		return nil, err
	}
	s.log.InfoContext(ctx, "idle users per team fetched", slog.Int("teams_count", len(metrics)))
	return metrics, nil
}

// GetNeedyPRsPerTeam returns the count of open PRs that need more reviewers, grouped by author's team.
func (s *PRService) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	metrics, err := s.prRepo.GetNeedyPRsPerTeam(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get needy PRs per team", slog.String("error", err.Error()))
		return nil, err
	}
	s.log.InfoContext(ctx, "needy PRs per team fetched", slog.Int("teams_count", len(metrics)))
	return metrics, nil
}
