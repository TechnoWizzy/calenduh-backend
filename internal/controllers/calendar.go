package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"calenduh-backend/internal/util"
	"errors"
	"github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
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

	if calendar.IsWebBased {
		_, found := util.WebCalendars.Get(calendar.CalendarID)
		if !found {
			cal, err := ics.ParseCalendarFromUrl(*calendar.Url)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			parsedCalendar, err := SaveICal(c, cal, true, calendar.Url)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			calendar = *parsedCalendar
		}
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
	cal.SetXWRCalID(calendar.CalendarID)
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

	for i, calendar := range calendars {
		if calendar.IsWebBased {
			_, found := util.WebCalendars.Get(calendar.CalendarID)
			if !found {
				cal, err := ics.ParseCalendarFromUrl(*calendar.Url)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				parsedCalendar, err := SaveICal(c, cal, true, calendar.Url)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				calendars[i] = *parsedCalendar
			}
		}
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

func ImportICal(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close file"})
		}
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	cal, err := ics.ParseCalendar(strings.NewReader(string(data)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ical format"})
		return
	}

	calendar, err := SaveICal(c, cal, false, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, calendar)
}

func SubscribeICal(c *gin.Context) {
	type SubscribeICalParams struct {
		Url string `json:"url"`
	}

	var params SubscribeICalParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cal, err := ics.ParseCalendarFromUrl(params.Url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	calendar, err := SaveICal(c, cal, true, &params.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.PureJSON(http.StatusOK, calendar)
}

func SaveICal(c *gin.Context, cal *ics.Calendar, isWebBased bool, url *string) (*sqlc.Calendar, error) {
	user := *ParseUser(c)
	var calName, calID string

	for _, prop := range cal.CalendarProperties {
		switch prop.IANAToken {
		case "X-WR-CALNAME":
			calName = prop.Value
		case "X-CALENDAR-ID":
			calID = prop.Value
		}
	}
	if calName == "" {
		calName = "Imported Calendar"
	}
	if calID == "" {
		// Generate or require a calendar ID
		calID = uuid.New().String()
	}

	if isWebBased {
		util.WebCalendars.Set(calID, cal, 0)
	}

	_ = database.Db.Queries.DeleteCalendar(c, calID) // Delete old calendar if it exists somehow
	calendar, err := database.Db.Queries.CreateCalendar(c, sqlc.CreateCalendarParams{
		CalendarID: calID,
		UserID:     &user.UserID,
		Title:      calName,
		Color:      "#4285F4",
		IsImported: !isWebBased,
		IsWebBased: isWebBased,
		IsPublic:   false,
		Url:        url,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("%d events on calendar\n", len(cal.Events()))
	createdEvents := 0
	for _, e := range cal.Events() {

		start, err := e.GetStartAt()
		if err != nil {
			start, err = e.GetAllDayStartAt()
			if err != nil {
				log.Printf("Error getting start time: %s\n", err.Error())
				continue
			}
		}
		end, err := e.GetEndAt()
		if err != nil {
			end, err = e.GetAllDayEndAt()
			if err != nil {
				log.Printf("Error getting end time: %s\n", err.Error())
				continue
			}
		}
		desc := e.GetProperty(ics.ComponentPropertyDescription)
		loc := e.GetProperty(ics.ComponentPropertyLocation)
		priority := e.GetProperty(ics.ComponentPropertyPriority)
		allDay := e.GetProperty(ics.ComponentPropertyDtStart).Value == "DATE"

		var descPtr, locPtr *string
		if desc != nil {
			d := desc.Value
			descPtr = &d
		}
		if loc != nil {
			l := loc.Value
			locPtr = &l
		}
		var priorityPtr *int32
		if priority != nil {
			p, _ := strconv.Atoi(priority.Value)
			p32 := int32(p)
			priorityPtr = &p32
		} else {
			val := int32(0)
			priorityPtr = &val
		}

		eventID := e.GetProperty(ics.ComponentPropertyUniqueId).Value
		name := e.GetProperty(ics.ComponentPropertySummary).Value

		log.Printf("ID: %s\n", eventID)
		log.Printf("Name: %s\n", name)
		log.Printf("Desc: %s\n", desc)
		log.Printf("Event: %s\n", eventID)

		_ = database.Db.Queries.DeleteEvent(c, eventID)
		_, err = database.Db.Queries.CreateEvent(c, sqlc.CreateEventParams{
			EventID:     eventID,
			CalendarID:  calendar.CalendarID,
			Name:        name,
			Description: descPtr,
			Location:    locPtr,
			StartTime:   start,
			EndTime:     end,
			AllDay:      allDay,
			Priority:    priorityPtr,
		})
		if err != nil {
		} else {
			createdEvents++
		}
	}
	log.Printf("%d events created\n", createdEvents)

	return &calendar, nil
}
