package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"strings"
	"time"
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

func GetAllPublicCalendars(c *gin.Context) {
	calendars, err := database.Db.Queries.GetAllPublicCalendars(c)
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

	if strings.HasSuffix(calendarId, ".ical") {
		GetCalendarICal(c, strings.TrimSuffix(calendarId, ".ical"))
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

func GetCalendarICal(c *gin.Context, calendarId string) {
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

	events, err := database.Db.Queries.GetEventsByCalendarId(c, sqlc.GetEventsByCalendarIdParams{
		CalendarID: calendar.CalendarID,
		EndTime:    time.UnixMilli(1 << 48),
	})
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "calendar events not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription("Generated Calendar: " + calendar.Title)
	cal.SetProductId("Calenduh Services 2025")

	for _, event := range events {
		icalEvent := cal.AddEvent(event.EventID)
		icalEvent.SetSummary(event.Name)
		icalEvent.SetColor(calendar.Color)
		icalEvent.SetDtStampTime(time.Now().UTC())

		if event.Description != nil && *event.Description != "" {
			icalEvent.SetDescription(*event.Description)
		}

		if event.Location != nil && *event.Location != "" {
			icalEvent.SetLocation(*event.Location)
		}

		icalEvent.SetStartAt(event.StartTime)
		icalEvent.SetEndAt(event.EndTime)

		if event.AllDay {
			icalEvent.SetAllDayStartAt(event.StartTime)
			icalEvent.SetAllDayEndAt(event.EndTime)
		}

		if event.Priority != nil {
			icalEvent.SetPriority(int(*event.Priority))
		}
	}

	data := cal.Serialize(ics.WithNewLine("\r\n"))
	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\"calendar.ics\"")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate") // Prevent aggressive caching
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.String(http.StatusOK, data)
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
	groupId := c.Param("group_id")
	if groupId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}
	for _, group := range groups {
		if groupId == group.GroupID {
			calendars, err := database.Db.Queries.GetCalendarsByGroupId(c, &group.GroupID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, calendars)
			return
		}
	}

	c.AbortWithStatus(http.StatusUnauthorized)
}

func GetAllGroupCalendars(c *gin.Context) {
	groups := *ParseGroups(c)
	calendars := make([]sqlc.Calendar, 0)
	for _, group := range groups {
		groupCalendars, err := database.Db.Queries.GetCalendarsByGroupId(c, &group.GroupID)
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
	input.GroupID = nil

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
	input.UserID = nil

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

func DeleteAllCalendars(c *gin.Context) {
	err := database.Db.Queries.DeleteAllCalendars(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "all calendars deleted successfully"})
}

func CanEditCalendar(calendar sqlc.Calendar, userId string, groups []sqlc.Group) bool {
	if calendar.UserID != nil && *calendar.UserID == userId {
		return true
	} else {
		for _, group := range groups {
			if calendar.GroupID != nil && *calendar.GroupID == group.GroupID {
				return true
			}
		}
	}

	return false
}
