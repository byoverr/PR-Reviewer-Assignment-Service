package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/apperrors"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/models"
	"github.com/byoverr/PR-Reviewer-Assignment-Service/internal/services"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	svc *services.TeamService
	log *slog.Logger
}

func NewTeamHandler(svc *services.TeamService, log *slog.Logger) *TeamHandler {
	return &TeamHandler{svc: svc, log: log}
}

// CreateTeam handles POST /team/add.
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req models.Team
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid create team request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	team, err := h.svc.CreateTeam(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("create team failed", slog.String("team_name", req.Name), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
}

// GetTeam handles GET /team/get?team_name=...
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		h.log.Warn("missing team_name query param")
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	teamName, _ = url.QueryUnescape(teamName)

	team, err := h.svc.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		h.log.Error("get team failed", slog.String("team_name", teamName), slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}

// AddMemberToTeam handles POST /team/add-member.
func (h *TeamHandler) AddMemberToTeam(c *gin.Context) {
	var req struct {
		TeamName string            `json:"team_name" binding:"required"`
		Member   models.TeamMember `json:"member" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("invalid add member request", slog.String("error", err.Error()))
		h.mapErrorToResponse(c, apperrors.ErrInvalidInput)
		return
	}

	err := h.svc.AddMemberToTeam(c.Request.Context(), req.TeamName, req.Member)
	if err != nil {
		h.log.Error("add member failed",
			slog.String("team_name", req.TeamName),
			slog.String("user_id", req.Member.UserID),
			slog.String("error", err.Error()))
		h.mapErrorToResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member added successfully"})
}

func (h *TeamHandler) mapErrorToResponse(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := ErrorCodeInternalError
	msg := ErrorMessageInternalError

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		status = http.StatusNotFound
		code = ErrorCodeNotFound
		msg = ErrorMessageNotFound
	case errors.Is(err, apperrors.ErrTeamExists):
		status = http.StatusBadRequest
		code = "TEAM_EXISTS"
		msg = "Team already exists"
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = http.StatusBadRequest
		code = ErrorCodeInvalidInput
		msg = ErrorMessageInvalidInput
	default:
		h.log.Error("unexpected error", slog.String("error", err.Error()))
	}

	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}
