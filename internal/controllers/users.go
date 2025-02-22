package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetMe
// @Summary Get details of the current user
// @Description Fetches the user data for the currently authenticated user.
func GetMe(c *gin.Context, user *sqlc.User, _ *[]sqlc.Group) {
	c.PureJSON(http.StatusOK, user)
	return
}

func GetAllUsers(c *gin.Context) {
	users, err := database.Queries.GetAllUsers(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, users)
}
