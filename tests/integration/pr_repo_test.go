package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig("postgres://user:pass@localhost:5433/pr_service_e2e?sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `TRUNCATE TABLE pull_requests, users, teams CASCADE`)
	require.NoError(t, err)

	return pool
}

func TestPRRepo_CreatePR(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u1", "User1", "team1", true)
		require.NoError(t, err)

		pr := &models.PullRequest{
			ID:        "pr-1",
			Title:     "Test PR",
			AuthorID:  "u1",
			Status:    "OPEN",
			Reviewers: []string{"u2"},
		}
		err = repo.CreatePR(ctx, pr)
		require.NoError(t, err)

		exists, err := repo.ExistsPR(ctx, "pr-1")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("PRExists", func(t *testing.T) {
		pr := &models.PullRequest{
			ID:       "pr-1",
			Title:    "Test PR",
			AuthorID: "u1",
			Status:   "OPEN",
		}
		err := repo.CreatePR(ctx, pr)
		assert.ErrorIs(t, err, apperrors.ErrPRExists)
	})
}

func TestPRRepo_GetPRByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u1", "User1", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			"pr-2", "Test PR", "u1", "OPEN", []string{"u2"}, now)
		require.NoError(t, err)

		pr, err := repo.GetPRByID(ctx, "pr-2")
		require.NoError(t, err)
		assert.Equal(t, "pr-2", pr.ID)
		assert.Equal(t, "Test PR", pr.Title)
		assert.Equal(t, "OPEN", pr.Status)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.GetPRByID(ctx, "pr-nonexist")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestPRRepo_UpdatePR(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u1", "User1", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			"pr-3", "Test PR", "u1", "OPEN", []string{"u2"}, now)
		require.NoError(t, err)

		pr := &models.PullRequest{
			ID:                "pr-3",
			Reviewers:         []string{"u3"},
			NeedMoreReviewers: false,
		}
		err = repo.UpdatePR(ctx, pr)
		require.NoError(t, err)

		updated, err := repo.GetPRByID(ctx, "pr-3")
		require.NoError(t, err)
		assert.Equal(t, []string{"u3"}, updated.Reviewers)
	})

	t.Run("PRMerged", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4) 
			ON CONFLICT (id) DO NOTHING`,
			"u1-merged", "User1 Merged", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			"pr-4", "Test PR", "u1-merged", "MERGED", []string{"u2"}, now)
		require.NoError(t, err)

		pr := &models.PullRequest{
			ID:        "pr-4",
			Reviewers: []string{"u3"},
		}
		err = repo.UpdatePR(ctx, pr)
		assert.ErrorIs(t, err, apperrors.ErrPRMerged)
	})
}

func TestPRRepo_MergePR(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4) 
			ON CONFLICT (id) DO NOTHING`,
			"u1-merge-test", "User1 Merge Test", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, reviewers, created_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			"pr-5", "Test PR", "u1-merge-test", "OPEN", []string{"u2"}, now)
		require.NoError(t, err)

		err = repo.MergePR(ctx, "pr-5")
		require.NoError(t, err)

		pr, err := repo.GetPRByID(ctx, "pr-5")
		require.NoError(t, err)
		assert.Equal(t, "MERGED", pr.Status)
		assert.NotNil(t, pr.MergedAt)
	})
}

func TestPRRepo_GetTotalPRs(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u1", "User1", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, created_at) 
			VALUES ($1, $2, $3, $4, $5)`,
			"pr-6", "PR1", "u1", "OPEN", now)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, created_at) 
			VALUES ($1, $2, $3, $4, $5)`,
			"pr-7", "PR2", "u1", "OPEN", now)
		require.NoError(t, err)

		total, err := repo.GetTotalPRs(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
	})
}

func TestPRRepo_GetPrsByStatus(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u1", "User1", "team1", true)
		require.NoError(t, err)

		now := time.Now()
		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, created_at) 
			VALUES ($1, $2, $3, $4, $5)`,
			"pr-8", "PR1", "u1", "OPEN", now)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO pull_requests (id, title, author_id, status, created_at, merged_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			"pr-9", "PR2", "u1", "MERGED", now, now)
		require.NoError(t, err)

		open, merged, err := repo.GetPrsByStatus(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, open, 1)
		assert.GreaterOrEqual(t, merged, 1)
	})
}

func TestPRRepo_Metrics(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewPRRepo(pool)
	ctx := context.Background()

	_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1), ($2)`, "team1", "team2")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES 
		('u1', 'User1', 'team1', true),
		('u2', 'User2', 'team1', true),
		('u3', 'User3', 'team2', true),
		('u4', 'User4', 'team2', false)`)
	require.NoError(t, err)

	now := time.Now()
	_, err = pool.Exec(ctx,
		`INSERT INTO pull_requests (id, title, author_id, status, reviewers, created_at, merged_at, need_more_reviewers)
		VALUES 
			('pr-10', 'PR1', 'u1', 'MERGED', ARRAY['u2'], $1, $2, false),
			('pr-11', 'PR2', 'u2', 'OPEN', ARRAY['u3'], $1, NULL, true),
			('pr-12', 'PR3', 'u3', 'OPEN', ARRAY['u1'], $1, NULL, false)`,
		now.Add(-2*time.Hour), now.Add(-time.Hour))
	require.NoError(t, err)

	t.Run("GetAvgCloseTime", func(t *testing.T) {
		avg, count, avgErr := repo.GetAvgCloseTime(ctx)
		require.NoError(t, avgErr)
		assert.Greater(t, avg, 0.0)
		assert.Equal(t, 1, count)
	})

	t.Run("GetIdleUsersPerTeam", func(t *testing.T) {
		metrics, idleErr := repo.GetIdleUsersPerTeam(ctx)
		require.NoError(t, idleErr)
		assert.Len(t, metrics, 1)
	})

	t.Run("GetNeedyPRsPerTeam", func(t *testing.T) {
		metrics, needyErr := repo.GetNeedyPRsPerTeam(ctx)
		require.NoError(t, needyErr)
		require.Len(t, metrics, 1)
		assert.Equal(t, "team1", metrics[0].TeamName)
		assert.Equal(t, 1, metrics[0].Count)
	})

	t.Run("GetAssignmentsPerUser", func(t *testing.T) {
		assignments, assignErr := repo.GetAssignmentsPerUser(ctx)
		require.NoError(t, assignErr)
		require.Len(t, assignments, 3)
		for _, a := range assignments {
			switch a.UserID {
			case "u1":
				assert.Equal(t, 1, a.Count)
			case "u2":
				assert.Equal(t, 1, a.Count)
			case "u3":
				assert.Equal(t, 1, a.Count)
			default:
				t.Errorf("unexpected user %s", a.UserID)
			}
		}
	})
}
