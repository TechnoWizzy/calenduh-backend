package controllers

import (
	"calenduh-backend/internal/database"
	"github.com/gin-gonic/gin"
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
	eventId := c.Query("event_id")
	event, err := database.FetchEventById(c, eventId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch event"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func FetchCalendarEvents(c *gin.Context) {
	calendarId := c.Query("calendar_id")
	events, err := database.FetchEventsByCalendarId(c, calendarId)
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
