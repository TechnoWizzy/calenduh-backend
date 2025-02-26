package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
)

func GetAllCalendars(c *gin.Context) {
	calendars, err := database.Db.Queries.GetAllCalendars(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Calendar, 0))
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendars)
}

func GetCalendar(c *gin.Context) {
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	calendar, err := database.Db.Queries.GetCalendarById(c, calendarId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendar)
}

func GetUserCalendars(c *gin.Context) {
	user := *ParseUser(c)
	calendars, err := database.Db.Queries.GetCalendarsByUserId(c, &user.UserID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Calendar, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendars)
}

func GetGroupCalendars(c *gin.Context) {
	groups := *ParseGroups(c)
	calendars := make([]sqlc.Calendar, 0)
	for _, group := range groups {
		groupCalendars, err := database.Db.Queries.GetCalendarsByUserId(c, &group.GroupID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		calendars = append(calendars, groupCalendars...)
	}

	c.JSON(http.StatusOK, calendars)
}

func GetSubscribedCalendars(c *gin.Context) {
	user := *ParseUser(c)
	calendars, err := database.Db.Queries.GetSubscribedCalendars(c, user.UserID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Calendar, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendars)
}

func CreateUserCalendar(c *gin.Context) {
	user := *ParseUser(c)
	var input sqlc.CreateCalendarParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	input.CalendarID = gonanoid.Must()
	input.UserID = &user.UserID

	calendar, err := database.Db.Queries.CreateCalendar(c, input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

func CreateGroupCalendar(c *gin.Context) {
	groups := *ParseGroups(c)
	groupId := c.Param("group_id")
	if groupId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	var input sqlc.CreateCalendarParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	input.CalendarID = gonanoid.Must()
	input.GroupID = &groupId

	if !CanEditGroup(*input.GroupID, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	calendar, err := database.Db.Queries.CreateCalendar(c, input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

func UpdateCalendar(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	var input sqlc.UpdateCalendarParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	input.CalendarID = calendarId
	input.UserID = &user.UserID

	calendar, err := database.Db.Queries.GetCalendarById(c, input.CalendarID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if !CanEditCalendar(calendar, user.UserID, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	calendar, err = database.Db.Queries.UpdateCalendar(c, input)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "event not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, calendar)
}

func DeleteCalendar(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	calendar, err := database.Db.Queries.GetCalendarById(c, calendarId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if !CanEditCalendar(calendar, user.UserID, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if err = database.Db.Queries.DeleteCalendar(c, calendarId); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete calendar" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "calendar deleted successfully"})
}

func CanEditCalendar(calendar sqlc.Calendar, userId string, groups []sqlc.Group) bool {
	if *calendar.UserID == userId {
		return true
	} else {
		for _, group := range groups {
			if *calendar.GroupID == group.GroupID {
				return true
			}
		}
	}

	return false
}
