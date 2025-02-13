package handlers

import (
	"github.com/gin-gonic/gin"
	"calenduh-backend/internal/database"
	"net/http"
)

func CreateEvent(c *gin.Context) {
	var options database.CreateEventOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := database.CreateEvent(c, &options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func FetchEvent(c *gin.Context) {
	id := c.Query("id")
	event, err := database.FetchEventById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch event"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func FetchEventName(c *gin.Context) {
	id := c.Query("id")
	event, err := database.FetchEventNameById(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch event name"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func FetchAllEvents(c *gin.Context) {
	events, err := database.FetchAllEvents(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch all events"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func UpdateEvent(c *gin.Context) {
	id := c.Query("id")
	var options database.UpdateEventOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	event, err := database.UpdateEvent(c, id, &options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update event"})
		return
	}
	c.JSON(http.StatusOK, event)
}