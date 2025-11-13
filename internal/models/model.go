package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
	TeamID   int    `json:"teamId"`
}

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PullRequest struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	AuthorID          int        `json:"authorId"`
	Status            string     `json:"status"` // OPEN / MERGED
	Reviewers         []int      `json:"reviewers"`
	NeedMoreReviewers bool       `json:"needMoreReviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}
