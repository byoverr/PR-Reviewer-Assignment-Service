-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
                       id SERIAL PRIMARY KEY,
                       name TEXT UNIQUE NOT NULL
);

CREATE TABLE users (
                       id TEXT PRIMARY KEY,
                       name TEXT NOT NULL,
                       team_name TEXT REFERENCES teams(name) ON DELETE SET NULL,
                       is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE pull_requests (
                               id TEXT PRIMARY KEY,
                               title TEXT NOT NULL,
                               author_id TEXT REFERENCES users(id),
                               status TEXT DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
                               reviewers TEXT[],
                               need_more_reviewers BOOLEAN DEFAULT FALSE,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               merged_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_reviewers ON pull_requests USING GIN(reviewers);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers;
DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_author_id;
DROP INDEX IF EXISTS idx_users_team_name;

DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

DROP TYPE IF EXISTS pr_status;
-- +goose StatementEnd
