package models

import "time"

type TeamMember struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	Name    string       `json:"team_name" binding:"required,min=1"`
	Members []TeamMember `json:"members" binding:"required,dive"`
}

type User struct {
	ID       string `json:"user_id" binding:"required"`
	Name     string `json:"username" binding:"required"`
	TeamName string `json:"team_name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	ID                string     `json:"pull_request_id" binding:"required"`
	Title             string     `json:"pull_request_name" binding:"required"`
	AuthorID          string     `json:"author_id" binding:"required"`
	Status            string     `json:"status" binding:"required,oneof=OPEN MERGED"`
	Reviewers         []string   `json:"assigned_reviewers" binding:"max=2"` // 0,1,2
	NeedMoreReviewers bool       `json:"need_more_reviewers,omitempty"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Title    string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}
