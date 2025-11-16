package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

var _ TeamRepository = (*TeamRepo)(nil)

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

// CreateTeam creates team.
func (r *TeamRepo) CreateTeam(ctx context.Context, team *models.Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to begin tx")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`, team.Name).Scan(&exists)
	if err != nil {
		return apperrors.Wrap(err, "failed to check team existence")
	}
	if exists {
		return apperrors.ErrTeamExists // ← Early return, no upsert
	}

	_, err = tx.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, team.Name)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert team")
	}

	for _, m := range team.Members {
		_, err = tx.Exec(ctx, `
			INSERT INTO users (id, name, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET name = $2, team_name = $3, is_active = $4
		`, m.UserID, m.Username, team.Name, m.IsActive)
		if err != nil {
			return apperrors.Wrap(err, "failed to upsert member")
		}
	}

	return nil
}

// GetTeamByName gets team by name.
func (r *TeamRepo) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	team := &models.Team{Name: name}

	err := r.db.QueryRow(ctx, `SELECT name FROM teams WHERE name = $1`, name).Scan(&team.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to query team")
	}

	rows, err := r.db.Query(ctx, `SELECT id, name, is_active FROM users WHERE team_name = $1`, name)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query members")
	}
	defer rows.Close()

	team.Members = []models.TeamMember{}
	for rows.Next() {
		var m models.TeamMember
		if scanErr := rows.Scan(&m.UserID, &m.Username, &m.IsActive); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan member")
		}
		team.Members = append(team.Members, m)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating members")
	}

	return team, nil
}
