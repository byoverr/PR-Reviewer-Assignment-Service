package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all API routes.
func SetupRoutes(r *gin.Engine, prHandler *PRHandler, teamHandler *TeamHandler, userHandler *UserHandler) {
	api := r.Group("/")

	// Teams
	api.POST("/team/add", teamHandler.CreateTeam)
	api.POST("/team/add-member", teamHandler.AddMemberToTeam) // New
	api.GET("/team/get", teamHandler.GetTeam)

	// Users
	api.POST("/users/setIsActive", userHandler.SetUserActive)
	api.GET("/users/getReview", userHandler.GetPRsForUser)
	api.POST("/users/deactivateByTeam", userHandler.DeactivateUsersByTeam)

	// PullRequests
	api.POST("/pullRequest/create", prHandler.CreatePR)
	api.POST("/pullRequest/merge", prHandler.MergePR)
	api.POST("/pullRequest/reassign", prHandler.ReassignReviewer)

	// Stats
	stats := api.Group("/stats")
	stats.GET("/prs-total", prHandler.GetTotalPRs)
	stats.GET("/prs-status", prHandler.GetPrsByStatus)
	stats.GET("/assignments-per-user", prHandler.GetAssignmentsPerUser)
	stats.GET("/top-reviewers", prHandler.GetTopReviewers)
	stats.GET("/avg-close-time", prHandler.GetAvgCloseTime)
	stats.GET("/idle-users-per-team", prHandler.GetIdleUsersPerTeam)
	stats.GET("/needy-prs-per-team", prHandler.GetNeedyPRsPerTeam)

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
