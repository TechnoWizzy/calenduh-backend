package handlers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateEventHandler(c *gin.Context) {
	var event sqlc.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := database.CreateEvent(context.Background(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Event created successfully!"})
}

func GetEventHandler(c *gin.Context) {
	eventID := c.Param("id")
	event, err := database.GetEvent(context.Background(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func UpdateEventHandler(c *gin.Context) {
	eventID := c.Param("id")
	var updatedEvent sqlc.Event

	if err := c.ShouldBindJSON(&updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	event, err := database.GetEvent(context.Background(), eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	event.CalendarID = updatedEvent.CalendarID
	event.Name = updatedEvent.Name
	event.Location = updatedEvent.Location
	event.Description = updatedEvent.Description
	event.Notification = updatedEvent.Notification
	event.Frequency = updatedEvent.Frequency
	event.Priority = updatedEvent.Priority
	event.StartTime = updatedEvent.StartTime
	event.EndTime = updatedEvent.EndTime

	if err := database.UpdateEvent(context.Background(), *event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func DeleteEventHandler(c *gin.Context) {
	eventID := c.Param("id")
	if err := database.DeleteEvent(context.Background(), eventID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
