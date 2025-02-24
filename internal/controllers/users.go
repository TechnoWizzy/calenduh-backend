package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
)

// GetMe
// @Summary Get details of the current user
// @Description Fetches the user data for the currently authenticated user.
func GetMe(c *gin.Context, user sqlc.User, _ []sqlc.Group) {
	c.PureJSON(http.StatusOK, user)
}

// GetAllUsers
// @Summary Lists all users in the database
// @Description Debug route to see all users currently existing
func GetAllUsers(c *gin.Context) {
	users, err := database.Db.Queries.GetAllUsers(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.User, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, users)
}

// GetUser
// @Summary Gets a User by ID
// @Description Used to retrieve user details for an account other than the one logged in
func GetUser(c *gin.Context, _ sqlc.User, _ []sqlc.Group) {
	userId := c.Param("user_id")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	user, err := database.Db.Queries.GetUserById(c, userId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context, user sqlc.User, _ []sqlc.Group) {
	var updateUserParams sqlc.UpdateUserParams
	if err := c.ShouldBindJSON(&updateUserParams); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateUserParams.UserID = user.UserID

	user, err := database.Db.Queries.UpdateUser(c, updateUserParams)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		}
	}
}

func DeleteMe(c *gin.Context, user sqlc.User, _ []sqlc.Group) {
	if err := database.Db.Queries.DeleteUser(c, user.UserID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"status": "deleted"})
}

func DeleteUser(c *gin.Context) {
	userId := c.Param("user_id")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if err := database.Db.Queries.DeleteUser(c, userId); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.PureJSON(http.StatusOK, gin.H{"status": "deleted"})
}
