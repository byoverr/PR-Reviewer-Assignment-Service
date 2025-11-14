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
		if err := rows.Scan(&p.ID, &p.Title, &p.AuthorID, &p.Status); err != nil {
			return nil, apperrors.Wrap(err, "failed to scan PR")
		}
		prs = append(prs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating PRs")
	}
	return prs, nil
}

func (r *PRRepo) ExistsPR(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check PR existence")
	}
	return exists, nil
}
