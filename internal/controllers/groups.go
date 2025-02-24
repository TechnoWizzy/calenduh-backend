package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
)

func GetAllGroups(c *gin.Context) {
	groups, err := database.Db.Queries.GetAllGroups(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "groups not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetGroup(c *gin.Context, _ sqlc.User, groups []sqlc.Group) {
	groupId := c.Param("group_id")
	if groupId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	for _, group := range groups {
		if group.GroupID == groupId {
			c.JSON(http.StatusOK, group)
			return
		}
	}

	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found or not permissible"})
}

func CreateGroup(c *gin.Context, user sqlc.User, groups []sqlc.Group) {
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		var input sqlc.CreateGroupParams
		if err := c.BindJSON(&input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		input.GroupID = gonanoid.Must()

		group, err := queries.CreateGroup(c, input)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}

		// ToDo Create Group Member

		c.JSON(http.StatusCreated, group)
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func UpdateGroup(c *gin.Context, _ sqlc.User, groups []sqlc.Group) {
	var input sqlc.UpdateGroupParams
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	input.GroupID = c.Param("group_id")
	if input.GroupID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
	}
	if !CanEditGroup(input.GroupID, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	group, err := database.Db.Queries.UpdateGroup(c, input)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

func DeleteGroup(c *gin.Context, _ sqlc.User, groups []sqlc.Group) {
	groupId := c.Param("group_id")
	if groupId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
	}

	if !CanEditGroup(groupId, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err := database.Db.Queries.DeleteGroup(c, groupId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func CanEditGroup(groupId string, groups []sqlc.Group) bool {
	for _, group := range groups {
		if group.GroupID == groupId {
			return true
		}
	}

	return false
}
