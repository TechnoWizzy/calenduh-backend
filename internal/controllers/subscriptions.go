package controllers

import (
	"calenduh-backend/internal/database"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateSubscription(c *gin.Context) {
	var options database.SubscribeOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := database.CreateSubscription(c, &options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

func DeleteSubscription(c *gin.Context) {
	v, found := c.Get("user")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user required"})
		return
	}

	user := v.(*database.User)
	calendarId := c.Param("calendar")
	if calendarId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "calendar id required"})
		return
	}

	err := database.DeleteSubscription(c, user.Id, calendarId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func FetchSubscriptions(c *gin.Context) {
	v, found := c.Get("user")
	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user required"})
		return
	}

	user := v.(*database.User)
	subscriptions, err := database.FetchAllSubscriptions(c, user.Id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to fetch subscribed calendars"})
		return
	}

	c.PureJSON(http.StatusOK, subscriptions)
}
