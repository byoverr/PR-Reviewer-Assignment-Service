package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
)

type UserRepo struct {
	db *pgxpool.Pool
}

var _ UserRepository = (*UserRepo)(nil)

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// UpsertUser creates or updates a user.
func (r *UserRepo) UpsertUser(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, name, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name = $2, team_name = $3, is_active = $4
	`, user.ID, user.Name, user.TeamName, user.IsActive)
	if err != nil {
		return apperrors.Wrap(err, "failed to upsert user")
	}
	return nil
}

// GetUserByID gets a user by ID.
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, team_name, is_active FROM users WHERE id = $1`,
		id).Scan(&u.ID, &u.Name, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, apperrors.Wrap(err, "failed to query user")
	}
	return u, nil
}

// UpdateUserActive updates the active status of a user.
func (r *UserRepo) UpdateUserActive(ctx context.Context, id string, isActive bool) error {
	res, err := r.db.Exec(ctx, `UPDATE users SET is_active = $2 WHERE id = $1`, id, isActive)
	if err != nil {
		return apperrors.Wrap(err, "failed to update user active")
	}
	if res.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// GetActiveUsersByTeam gets all active users for a team.
func (r *UserRepo) GetActiveUsersByTeam(ctx context.Context, teamName string) ([]models.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, team_name, is_active FROM users 
		WHERE team_name = $1 AND is_active = true
	`, teamName)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query active users")
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if scanErr := rows.Scan(&u.ID, &u.Name, &u.TeamName, &u.IsActive); scanErr != nil {
			return nil, apperrors.Wrap(scanErr, "failed to scan user")
		}
		users = append(users, u)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, apperrors.Wrap(scanErr, "error iterating users")
	}
	return users, nil
}

// GetTeamNameByUserID gets the team name for a user by user ID.
func (r *UserRepo) GetTeamNameByUserID(ctx context.Context, userID string) (string, error) {
	var teamName string
	err := r.db.QueryRow(ctx, `SELECT team_name FROM users WHERE id = $1`, userID).Scan(&teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrNotFound
		}
		return "", apperrors.Wrap(err, "failed to query team name")
	}
	return teamName, nil
}

// DeactivateUsersByTeam deactivates all users in a team.
func (r *UserRepo) DeactivateUsersByTeam(ctx context.Context, teamName string) error {
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

	_, err = tx.Exec(ctx, `UPDATE users SET is_active = false WHERE team_name = $1`, teamName)
	if err != nil {
		return apperrors.Wrap(err, "failed to deactivate users")
	}

	return nil
}
