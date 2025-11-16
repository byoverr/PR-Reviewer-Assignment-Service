package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

type PRRepo struct {
	db *pgxpool.Pool
}

var _ PRRepository = (*PRRepo)(nil)

func NewPRRepo(db *pgxpool.Pool) *PRRepo {
	return &PRRepo{db: db}
}

// CreatePR creates PR.
func (r *PRRepo) CreatePR(ctx context.Context, pr *models.PullRequest) error {
	exists, err := r.ExistsPR(ctx, pr.ID)
	if err != nil {
		return apperrors.Wrap(err, "failed to check PR existence")
	}
	if exists {
		return apperrors.ErrPRExists
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO pull_requests (id, title, author_id, status, reviewers, need_more_reviewers, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, pr.ID, pr.Title, pr.AuthorID, pr.Status, pr.Reviewers, pr.NeedMoreReviewers, time.Now())
	if err != nil {
		return apperrors.Wrap(err, "failed to create PR")
	}
	return nil
}

// GetPRByID gets PRs by ID.
func (r *PRRepo) GetPRByID(ctx context.Context, id string) (*models.PullRequest, error) {
	pr := &models.PullRequest{}
	var createdAt time.Time
	var mergedAt *time.Time
	err := r.db.QueryRow(ctx, `
		SELECT id, title, author_id, status, reviewers, need_more_reviewers, created_at, merged_at
		FROM pull_requests WHERE id = $1
	`, id).Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.Reviewers, &pr.NeedMoreReviewers, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to query PR")
	}
	pr.CreatedAt = &createdAt
	pr.MergedAt = mergedAt
	return pr, nil
}

// UpdatePR updates PR.
func (r *PRRepo) UpdatePR(ctx context.Context, pr *models.PullRequest) error {
	current, err := r.GetPRByID(ctx, pr.ID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get PR for update")
	}
	if current.Status == "MERGED" {
		return apperrors.ErrPRMerged
	}

	_, err = r.db.Exec(ctx, `
		UPDATE pull_requests SET reviewers = $2, need_more_reviewers = $3
		WHERE id = $1
	`, pr.ID, pr.Reviewers, pr.NeedMoreReviewers)
	if err != nil {
		return apperrors.Wrap(err, "failed to update PR")
	}
	return nil
}

// MergePR merge PR idempotently.
func (r *PRRepo) MergePR(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE pull_requests SET status = 'MERGED', merged_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status != 'MERGED'
	`, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to merge PR")
	}
	return nil
}

// GetPRsForUser gets PRs for user.
func (r *PRRepo) GetPRsForUser(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, author_id, status
		FROM pull_requests 
		WHERE $1 = ANY(reviewers)
	`, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query PRs for user")
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var p models.PullRequestShort
		if scanErr := rows.Scan(&p.ID, &p.Title, &p.AuthorID, &p.Status); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan PR")
		}
		prs = append(prs, p)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating PRs")
	}
	return prs, nil
}

// ExistsPR checks pull requests for existence.
func (r *PRRepo) ExistsPR(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check PR existence")
	}
	return exists, nil
}

// GetTotalPRs returns the total count of all pull requests.
func (r *PRRepo) GetTotalPRs(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests`).Scan(&total)
	if err != nil {
		return 0, apperrors.Wrap(err, "failed to count total PRs")
	}
	return total, nil
}

// GetPrsByStatus returns the count of open and merged pull requests.
func (r *PRRepo) GetPrsByStatus(ctx context.Context) (int, int, error) {
	var open, merged int
	err := r.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'OPEN'),
			COUNT(*) FILTER (WHERE status = 'MERGED')
		FROM pull_requests
	`).Scan(&open, &merged)
	if err != nil {
		return 0, 0, apperrors.Wrap(err, "failed to count PRs by status")
	}
	return open, merged, nil
}

// GetAssignmentsPerUser returns the number of PR assignments per active user, ordered by count descending.
func (r *PRRepo) GetAssignmentsPerUser(ctx context.Context) ([]models.UserAssignment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.name, COUNT(pr.id) as count
		FROM users u
		JOIN pull_requests pr ON u.id = ANY(pr.reviewers)
		WHERE u.is_active = true
		GROUP BY u.id, u.name
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to count assignments per user")
	}
	defer rows.Close()

	var assignments []models.UserAssignment
	for rows.Next() {
		var ua models.UserAssignment
		if scanErr := rows.Scan(&ua.UserID, &ua.Name, &ua.Count); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan assignment")
		}
		assignments = append(assignments, ua)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating assignments")
	}
	return assignments, nil
}

// GetTopReviewers returns the top 5 reviewers by assignment count.
func (r *PRRepo) GetTopReviewers(ctx context.Context) ([]models.UserAssignment, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.name, COUNT(pr.id) as count
		FROM users u
		JOIN pull_requests pr ON u.id = ANY(pr.reviewers)
		WHERE u.is_active = true
		GROUP BY u.id, u.name
		ORDER BY count DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get top reviewers")
	}
	defer rows.Close()

	var top []models.UserAssignment
	for rows.Next() {
		var ua models.UserAssignment
		if scanErr := rows.Scan(&ua.UserID, &ua.Name, &ua.Count); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan top reviewer")
		}
		top = append(top, ua)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating top reviewers")
	}
	return top, nil
}

// GetAvgCloseTime returns the average time in seconds to close merged PRs and the count of merged PRs.
func (r *PRRepo) GetAvgCloseTime(ctx context.Context) (float64, int, error) {
	var avgSeconds float64
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT 
			AVG(EXTRACT(epoch FROM (merged_at - created_at))),
			COUNT(*)
		FROM pull_requests
		WHERE status = 'MERGED'
	`).Scan(&avgSeconds, &count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, nil
		}
		return 0, 0, apperrors.Wrap(err, "failed to calculate avg close time")
	}
	return avgSeconds, count, nil
}

// GetIdleUsersPerTeam returns active users with 0 assignments, grouped by team.
func (r *PRRepo) GetIdleUsersPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.team_name, COUNT(u.id) as count
		FROM users u
		WHERE u.is_active = true
		  AND u.id NOT IN (
		    SELECT DISTINCT unnest(pr.reviewers)
		    FROM pull_requests pr
		    WHERE pr.status = 'OPEN'
		  )
		GROUP BY u.team_name
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get idle users per team")
	}
	defer rows.Close()

	var metrics []models.TeamMetric
	for rows.Next() {
		var tm models.TeamMetric
		if scanErr := rows.Scan(&tm.TeamName, &tm.Count); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan idle metric")
		}
		metrics = append(metrics, tm)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating idle metrics")
	}
	return metrics, nil
}

// GetNeedyPRsPerTeam returns OPEN PRs with need_more_reviewers=true, grouped by author's team.
func (r *PRRepo) GetNeedyPRsPerTeam(ctx context.Context) ([]models.TeamMetric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.team_name, COUNT(pr.id) as count
		FROM pull_requests pr
		JOIN users u ON pr.author_id = u.id
		WHERE pr.status = 'OPEN'
		  AND pr.need_more_reviewers = true
		GROUP BY u.team_name
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get needy PRs per team")
	}
	defer rows.Close()

	var metrics []models.TeamMetric
	for rows.Next() {
		var tm models.TeamMetric
		if scanErr := rows.Scan(&tm.TeamName, &tm.Count); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan needy metric")
		}
		metrics = append(metrics, tm)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating needy metrics")
	}
	return metrics, nil
}

// GetOpenPRsWithReviewersFromTeam returns all OPEN PRs that have reviewers from the specified team.
func (r *PRRepo) GetOpenPRsWithReviewersFromTeam(ctx context.Context, teamName string) ([]models.PullRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status, pr.reviewers, pr.need_more_reviewers, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN users u ON u.id = ANY(pr.reviewers)
		WHERE pr.status = 'OPEN'
		  AND u.team_name = $1
	`, teamName)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get open PRs with reviewers from team")
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		var createdAt time.Time
		var mergedAt *time.Time
		if scanErr := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.Reviewers, &pr.NeedMoreReviewers, &createdAt, &mergedAt); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan PR")
		}
		pr.CreatedAt = &createdAt
		pr.MergedAt = mergedAt
		prs = append(prs, pr)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating PRs")
	}
	return prs, nil
}
