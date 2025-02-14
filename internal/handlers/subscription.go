package handlers

import (
	"github.com/gin-gonic/gin"
	"calenduh-backend/internal/database"
	"net/http"
)

func Subscribe(c *gin.Context) {
	var options database.SubscribeOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub, err := database.Subscribe(c, &options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sub)
}


func Unsubscribe(c *gin.Context) {
	var options database.UnsubscribeOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sub, err := database.Unsubscribe(c, &options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func FetchAllSubbedCalendars(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}
	subs, err := database.FetchAllSubscriptions(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch subscribed calendars"})
		return
	}
	c.JSON(http.StatusOK, subs)
}