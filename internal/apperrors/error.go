package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")           // NOT_FOUND
	ErrTeamExists   = errors.New("team already exists")          // TEAM_EXISTS
	ErrPRExists     = errors.New("PR already exists")            // PR_EXISTS
	ErrPRMerged     = errors.New("cannot modify merged PR")      // PR_MERGED
	ErrNotAssigned  = errors.New("reviewer not assigned to PR")  // NOT_ASSIGNED
	ErrNoCandidate  = errors.New("no active candidates in team") // NO_CANDIDATE
	ErrInvalidInput = errors.New("invalid input")                // INVALID_INPUT
	ErrInternal     = errors.New("internal error")               // INTERNAL_ERROR
)

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return errors.Join(errors.New(msg), err)
}
