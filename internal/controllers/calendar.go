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
