package integration_test

import (
	"context"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ func(*testing.T) *pgxpool.Pool = setupTestDB

func TestUserRepo_UpsertUser(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	t.Run("Success_Insert", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)

		user := &models.User{
			ID:       "u1",
			Name:     "User1",
			TeamName: "team1",
			IsActive: true,
		}
		err = repo.UpsertUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByID(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "u1", retrieved.ID)
		assert.Equal(t, "User1", retrieved.Name)
	})

	t.Run("Success_Update", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team2")
		require.NoError(t, err)

		user := &models.User{
			ID:       "u1",
			Name:     "User1 Updated",
			TeamName: "team2",
			IsActive: false,
		}
		err = repo.UpsertUser(ctx, user)
		require.NoError(t, err)

		retrieved, err := repo.GetUserByID(ctx, "u1")
		require.NoError(t, err)
		assert.Equal(t, "User1 Updated", retrieved.Name)
		assert.Equal(t, "team2", retrieved.TeamName)
		assert.False(t, retrieved.IsActive)
	})
}

func TestUserRepo_GetUserByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u2", "User2", "team1", true)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "u2")
		require.NoError(t, err)
		assert.Equal(t, "u2", user.ID)
		assert.Equal(t, "User2", user.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.GetUserByID(ctx, "u-nonexist")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestUserRepo_UpdateUserActive(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u3", "User3", "team1", true)
		require.NoError(t, err)

		err = repo.UpdateUserActive(ctx, "u3", false)
		require.NoError(t, err)

		user, err := repo.GetUserByID(ctx, "u3")
		require.NoError(t, err)
		assert.False(t, user.IsActive)
	})

	t.Run("NotFound", func(t *testing.T) {
		err := repo.UpdateUserActive(ctx, "u-nonexist", true)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}

func TestUserRepo_GetActiveUsersByTeam(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team1")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u4", "User4", "team1", true)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u5", "User5", "team1", true)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u6", "User6", "team1", false)
		require.NoError(t, err)

		users, err := repo.GetActiveUsersByTeam(ctx, "team1")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)
		for _, u := range users {
			assert.True(t, u.IsActive)
			assert.Equal(t, "team1", u.TeamName)
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		users, err := repo.GetActiveUsersByTeam(ctx, "team-nonexist")
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepo_DeactivateUsersByTeam(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewUserRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT DO NOTHING`, "team2")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u7", "User7", "team2", true)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u8", "User8", "team2", true)
		require.NoError(t, err)

		err = repo.DeactivateUsersByTeam(ctx, "team2")
		require.NoError(t, err)

		users, err := repo.GetActiveUsersByTeam(ctx, "team2")
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}
