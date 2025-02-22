package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FetchEvent(c *gin.Context) {
	eventId := c.Param("event_id")
	if eventId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	event, err := database.Queries.GetEvent(c, eventId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "event not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

func CreateEvent(c *gin.Context) {
	var input sqlc.CreateEventParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input" + err.Error()})
		return
	}

	err := database.Queries.CreateEvent(c, input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create event" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event created successfully"})
}

func UpdateEvent(c *gin.Context) {
	eventId := c.Param("event_id")
	if eventId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	var input sqlc.UpdateEventParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
 
	input.EventID = eventId

	if err := database.Queries.UpdateEvent(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event updated successfully"})
}

func deleteEvent(c *gin.Context) {
	eventId := c.Param("event_id")
	if eventId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	if err := database.Queries.DeleteEvent(c.Request.Context(), eventId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event deleted successfully"})
}