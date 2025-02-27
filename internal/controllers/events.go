package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"strconv"
	"time"
)

func GetAllEvents(c *gin.Context) {
	start, end := ParseRange(c)
	events, err := database.Db.Queries.GetAllEvents(c, sqlc.GetAllEventsParams{
		StartTime: *start,
		EndTime:   *end,
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Event, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, events)
}

func GetUserEvents(c *gin.Context) {
	user := *ParseUser(c)
	start, end := ParseRange(c)
	events, err := database.Db.Queries.GetEventsByUserId(c, sqlc.GetEventsByUserIdParams{
		UserID:    user.UserID,
		StartTime: *start,
		EndTime:   *end,
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Event, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, events)
}

func GetEvent(c *gin.Context) {
	calendarId := c.Param("calendar_id")
	eventId := c.Param("event_id")

	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}
	if eventId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	event, err := database.Db.Queries.GetEventById(c, eventId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "event not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

func GetCalendarEvents(c *gin.Context) {
	start, end := ParseRange(c)
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	events, err := database.Db.Queries.GetEventByCalendarId(c, sqlc.GetEventByCalendarIdParams{
		CalendarID: calendarId,
		StartTime:  *start,
		EndTime:    *end,
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Event, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, events)
}

func CreateEvent(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	calendarId := c.Param("calendar_id")
	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}

	var input sqlc.CreateEventParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	input.CalendarID = calendarId
	input.EventID = gonanoid.Must()

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

	event, err := database.Db.Queries.CreateEvent(c, input)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

func UpdateEvent(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	calendarId := c.Param("calendar_id")
	eventId := c.Param("event_id")

	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}
	if eventId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	var input sqlc.UpdateEventParams
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.CalendarID = calendarId
	input.EventID = eventId

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

	event, err := database.Db.Queries.UpdateEvent(c, input)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "event not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, event)
}

func DeleteEvent(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	calendarId := c.Param("calendar_id")
	eventId := c.Param("event_id")

	if calendarId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "calendar_id is required"})
		return
	}
	if eventId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	deleteEventParams := sqlc.DeleteEventParams{
		CalendarID: calendarId,
		EventID:    eventId,
	}

	calendar, err := database.Db.Queries.GetCalendarById(c, deleteEventParams.CalendarID)
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

	if err = database.Db.Queries.DeleteEvent(c.Request.Context(), deleteEventParams); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "event not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "event deleted successfully"})
}

func WithRange(c *gin.Context) {
	// Get start and end from query parameters
	startStr := c.Query("start")
	endStr := c.Query("end")

	// Initialize start and end times with min and max values
	minTime := time.UnixMilli(0)       // 1970-01-01 00:00:00 UTC
	maxTime := time.UnixMilli(1 << 48) // huge time

	start := minTime
	end := maxTime

	// Parse start time if provided
	if startStr != "" {
		startMs, err := strconv.ParseInt(startStr, 10, 64)
		if err == nil {
			startTime := time.UnixMilli(startMs)
			start = startTime
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}

	// Parse end time if provided
	if endStr != "" {
		endMs, err := strconv.ParseInt(endStr, 10, 64)
		if err == nil {
			endTime := time.UnixMilli(endMs)
			end = endTime
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}

	c.Set("start", start)
	c.Set("end", end)
}

func ParseRange(c *gin.Context) (*time.Time, *time.Time) {
	start := c.GetTime("start")
	end := c.GetTime("end")

	if start.After(end) {
		panic("start time must be before end time")
	}

	return &start, &end
}
