package models

import "time"

type TeamMember struct {
	UserID   string `json:"user_id"   binding:"required"`
	Username string `json:"username"  binding:"required"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	Name    string       `json:"team_name" binding:"required,min=1"`
	Members []TeamMember `json:"members"   binding:"required,dive"`
}

type User struct {
	ID       string `json:"user_id"   binding:"required"`
	Name     string `json:"username"  binding:"required"`
	TeamName string `json:"team_name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	ID                string     `json:"pull_request_id"               binding:"required"`
	Title             string     `json:"pull_request_name"             binding:"required"`
	AuthorID          string     `json:"author_id"                     binding:"required"`
	Status            string     `json:"status,omitempty"`
	Reviewers         []string   `json:"assigned_reviewers,omitempty"`
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

type PrsTotal struct {
	TotalPRs int `json:"total_prs"`
}

type PrsStatus struct {
	OpenPRs   int `json:"open_prs"`
	MergedPRs int `json:"merged_prs"`
}

type UserAssignment struct {
	UserID string `json:"user_id"`
	Name   string `json:"username"`
	Count  int    `json:"assignment_count"`
}

type UserAssignments struct {
	Assignments []UserAssignment `json:"user_assignments"`
}

type TopReviewers struct {
	TopReviewers []UserAssignment `json:"top_reviewers"`
}

type AvgCloseTimeDetail struct {
	AverageSeconds float64 `json:"average_seconds"`
	Breakdown      struct {
		Days    int `json:"days"`
		Hours   int `json:"hours"`
		Minutes int `json:"minutes"`
		Seconds int `json:"seconds"`
	} `json:"breakdown"`
	MergedPRsCount int `json:"merged_prs_count"`
}

type TeamMetric struct {
	TeamName string `json:"team_name"`
	Count    int    `json:"count"`
}

type TeamMetrics struct {
	Metrics []TeamMetric `json:"team_metrics"`
}
