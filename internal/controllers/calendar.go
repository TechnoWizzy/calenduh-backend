package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FetchCalendar(c *gin.Context) {
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	calendar, err := database.Queries.GetCalendarById(c, calendarId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendar)
}

func FetchUserCalendars(c *gin.Context, user *sqlc.User) {
	calendars, err := database.Queries.GetCalendarsByUserId(c, &user.UserID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no calendars found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendars)
}

func FetchGroupCalendars(c *gin.Context, user *sqlc.User) {

}

func FetchSubscribedCalendars(c *gin.Context) {

}

func CreateCalendar(c *gin.Context) {
	var input sqlc.CreateCalendarParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input" + err.Error()})
		return
	}

	err := database.Queries.CreateCalendar(c, input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "calendar created successfully"})
}

func DeleteCalendar(c *gin.Context) {
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	err := database.Queries.DeleteCalendar(c, calendarId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "calendar deleted successfully"})
}

func DeleteAllUserCalendars(c *gin.Context) {
	userId := c.Param("user_id")
	if userId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	err := database.Queries.DeleteAllUserCalendars(c, &userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete calendars" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "calendars deleted successfully"})
}

func UpdateCalendar(c *gin.Context) {
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	var input sqlc.UpdateCalendarParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input" + err.Error()})
		return
	}

	input.CalendarID = calendarId

	err := database.Queries.UpdateCalendar(c, input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "calendar updated successfully"})
}