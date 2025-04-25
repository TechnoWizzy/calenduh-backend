package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
)

func GetAllSubscriptions(c *gin.Context) {
	subscriptions, err := database.Db.Queries.GetAllSubscriptions(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Calendar, 0))
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

func CreateSubscription(c *gin.Context) {
	user := *ParseUser(c)
	var input sqlc.CreateSubscriptionParams
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.UserID = user.UserID

	if input.CalendarID == "" {
		if input.InviteCode == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing invite code"})
			return
		}
		_, err := database.Db.Queries.GetCalendarByInviteCode(c, *input.InviteCode)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
			return
		}
	}

	if err := database.Db.Queries.CreateSubscription(c, input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteSubscription(c *gin.Context) {
	userId := c.Param("user_id")
	calendarId := c.Param("calendar_id")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	if err := database.Db.Queries.DeleteSubscription(c, sqlc.DeleteSubscriptionParams{
		UserID:     userId,
		CalendarID: calendarId,
	}); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func DeleteMySubscription(c *gin.Context) {
	user := *ParseUser(c)
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	if err := database.Db.Queries.DeleteSubscription(c, sqlc.DeleteSubscriptionParams{
		UserID:     user.UserID,
		CalendarID: calendarId,
	}); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
