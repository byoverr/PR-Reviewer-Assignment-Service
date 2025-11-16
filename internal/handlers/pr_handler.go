package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"

	"github.com/gin-gonic/gin"
)

type PRHandler struct {
	svc *services.PRService
	log *slog.Logger
}

func NewPRHandler(svc *services.PRService, log *slog.Logger) *PRHandler {
	return &PRHandler{svc: svc, log: log}
}

// CreatePR handles POST /pullRequest/create.
func (h *PRHandler) CreatePR(c *gin.Context) {
	var req models.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid create PR request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	pr, err := h.svc.CreatePR(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("create PR failed", slog.String("pr_id", req.ID), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": pr})
}

// MergePR handles POST /pullRequest/merge.
func (h *PRHandler) MergePR(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid merge PR request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	pr, err := h.svc.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		h.log.Error("merge PR failed", slog.String("pr_id", req.PullRequestID), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr})
}

// ReassignReviewer handles POST /pullRequest/reassign.
func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req struct {
		PullRequestID string `json:"pull_request_id" binding:"required"`
		OldReviewerID string `json:"old_reviewer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid reassign request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	pr, newReviewer, err := h.svc.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		h.log.Error("reassign failed",
			slog.String("pr_id", req.PullRequestID),
			slog.String("old_reviewer", req.OldReviewerID),
			slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": pr, "replaced_by": newReviewer})
}

// GetTotalPRs handles GET /stats/total-prs.
func (h *PRHandler) GetTotalPRs(c *gin.Context) {
	total, err := h.svc.GetTotalPRs(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.PrsTotal{TotalPRs: total})
}

// GetPrsByStatus handles GET /stats/prs-by-status.
func (h *PRHandler) GetPrsByStatus(c *gin.Context) {
	open, merged, err := h.svc.GetPrsByStatus(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.PrsStatus{OpenPRs: open, MergedPRs: merged})
}

// GetAssignmentsPerUser handles GET /stats/assignments-per-user.
func (h *PRHandler) GetAssignmentsPerUser(c *gin.Context) {
	assignments, err := h.svc.GetAssignmentsPerUser(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.UserAssignments{Assignments: assignments})
}

// GetTopReviewers handles GET /stats/top-reviewers.
func (h *PRHandler) GetTopReviewers(c *gin.Context) {
	top, err := h.svc.GetTopReviewers(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.TopReviewers{TopReviewers: top})
}

// GetAvgCloseTime handles GET /stats/avg-close-time.
func (h *PRHandler) GetAvgCloseTime(c *gin.Context) {
	detail, err := h.svc.GetAvgCloseTime(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// GetIdleUsersPerTeam handles GET /stats/idle-users-per-team.
func (h *PRHandler) GetIdleUsersPerTeam(c *gin.Context) {
	metrics, err := h.svc.GetIdleUsersPerTeam(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.TeamMetrics{Metrics: metrics})
}

// GetNeedyPRsPerTeam handles GET /stats/needy-prs-per-team.
func (h *PRHandler) GetNeedyPRsPerTeam(c *gin.Context) {
	metrics, err := h.svc.GetNeedyPRsPerTeam(c.Request.Context())
	if err != nil {
		h.mapErrorToResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, models.TeamMetrics{Metrics: metrics})
}

func (h *PRHandler) mapErrorToResponse(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := ErrorCodeInternalError
	msg := ErrorMessageInternalError

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		code = ErrorCodeNotFound
		msg = ErrorMessageNotFound
	case errors.Is(err, apperrors.ErrPRExists):
		status = http.StatusConflict
		code = "PR_EXISTS"
		msg = "PR already exists"
	case errors.Is(err, apperrors.ErrPRMerged):
		status = http.StatusConflict
		code = "PR_MERGED"
		msg = "Cannot modify merged PR"
	case errors.Is(err, apperrors.ErrNotAssigned):
		status = http.StatusConflict
		code = "NOT_ASSIGNED"
		msg = "Reviewer not assigned"
	case errors.Is(err, apperrors.ErrNoCandidate):
		status = http.StatusConflict
		code = "NO_CANDIDATE"
		msg = "No available candidates"
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		code = ErrorCodeInvalidInput
		msg = ErrorMessageInvalidInput
	default:
		h.log.Error("unexpected error", slog.String("error", err.Error()))
	}

	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}
