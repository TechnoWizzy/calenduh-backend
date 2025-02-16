package controllers

import (
	"calenduh-backend/internal/database"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateGroup(c *gin.Context) {
	var options database.CreateGroupOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group, err := database.CreateGroup(c, &options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func FetchGroup(c *gin.Context) {
	id := c.Query("id")
	group, err := database.FetchGroupById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch group"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func FetchAllGroups(c *gin.Context) {
	groups, err := database.FetchAllGroups(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch all groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func UpdateGroup(c *gin.Context) {
	id := c.Query("id")
	var options database.UpdateGroupOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group, err := database.UpdateGroup(c, id, &options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update group"})
		return
	}
	c.JSON(http.StatusOK, group)
}
