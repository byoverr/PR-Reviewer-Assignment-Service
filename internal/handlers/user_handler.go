package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *services.UserService
	log *slog.Logger
}

func NewUserHandler(svc *services.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

// SetUserActive handles POST /users/setIsActive.
func (h *UserHandler) SetUserActive(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid set active request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	user, err := h.svc.SetUserActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		h.log.Error("set user active failed", slog.String("user_id", req.UserID), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GetPRsForUser handles GET /users/getReview?user_id=...
func (h *UserHandler) GetPRsForUser(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		h.log.Warn("missing user_id query param")
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	userID, _ = url.QueryUnescape(userID)

	prs, err := h.svc.GetPRsForUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("get PRs for user failed", slog.String("user_id", userID), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "pull_requests": prs})
}

// DeactivateUsersByTeam handles POST /users/deactivateByTeam.
func (h *UserHandler) DeactivateUsersByTeam(c *gin.Context) {
	var req struct {
		TeamName string `json:"team_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid deactivate by team request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	err := h.svc.DeactivateUsersByTeam(c.Request.Context(), req.TeamName)
	if err != nil {
		h.log.Error("deactivate users by team failed",
			slog.String("team_name", req.TeamName),
			slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "users deactivated and PRs reassigned successfully"})
}

func (h *UserHandler) mapErrorToResponse(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := ErrorCodeInternalError
	msg := ErrorMessageInternalError

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		code = ErrorCodeNotFound
		msg = ErrorMessageNotFound
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		code = ErrorCodeInvalidInput
		msg = ErrorMessageInvalidInput
	default:
		h.log.Error("unexpected error", slog.String("error", err.Error()))
	}

	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}
