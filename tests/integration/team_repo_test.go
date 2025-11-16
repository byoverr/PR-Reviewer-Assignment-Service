package integration_test

import (
	"context"
	"testing"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepo_CreateTeam(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewTeamRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u1", Username: "User1", IsActive: true},
				{UserID: "u2", Username: "User2", IsActive: true},
			},
		}
		err := repo.CreateTeam(ctx, team)
		require.NoError(t, err)

		retrieved, err := repo.GetTeamByName(ctx, "team1")
		require.NoError(t, err)
		assert.Equal(t, "team1", retrieved.Name)
		assert.Len(t, retrieved.Members, 2)
	})

	t.Run("TeamExists", func(t *testing.T) {
		team := &models.Team{
			Name: "team1",
			Members: []models.TeamMember{
				{UserID: "u3", Username: "User3", IsActive: true},
			},
		}
		err := repo.CreateTeam(ctx, team)
		assert.ErrorIs(t, err, apperrors.ErrTeamExists)
	})
}

func TestTeamRepo_GetTeamByName(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := repository.NewTeamRepo(pool)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		_, err := pool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, "team2")
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `INSERT INTO users (id, name, team_name, is_active) VALUES ($1, $2, $3, $4)`,
			"u4", "User4", "team2", true)
		require.NoError(t, err)

		team, err := repo.GetTeamByName(ctx, "team2")
		require.NoError(t, err)
		assert.Equal(t, "team2", team.Name)
		assert.Len(t, team.Members, 1)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := repo.GetTeamByName(ctx, "team-nonexist")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}
