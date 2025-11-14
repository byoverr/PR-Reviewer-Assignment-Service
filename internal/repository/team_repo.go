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

	// Insert team
	_, err = tx.Exec(ctx, `INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, team.Name)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert team")
	}

	// Check existence
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`, team.Name).Scan(&exists)
	if err != nil {
		return apperrors.Wrap(err, "failed to check team existence")
	}
	if !exists {
		return apperrors.ErrTeamExists
	}

	// Upsert members
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

func (r *TeamRepo) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	team := &models.Team{Name: name}

	// Check team
	err := r.db.QueryRow(ctx, `SELECT name FROM teams WHERE name = $1`, name).Scan(&team.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to query team")
	}

	// Get members
	rows, err := r.db.Query(ctx, `SELECT id, name, is_active FROM users WHERE team_name = $1`, name)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query members")
	}
	defer rows.Close()

	team.Members = []models.TeamMember{}
	for rows.Next() {
		var m models.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, apperrors.Wrap(err, "failed to scan member")
		}
		team.Members = append(team.Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating members")
	}

	return team, nil
}
