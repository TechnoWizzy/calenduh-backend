package database

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	EventId      string `json:"eventId" bson:"_id"`
	CalendarId   string `json:"calendarId" bson:"calendar_id"`
	Name         string `json:"name" bson:"name"`
	Start        int64  `json:"start" bson:"start"` // Start date and time
	End          int64  `json:"end" bson:"end"`     // End date and time
	Location     string `json:"location" bson:"location"`
	Description  string `json:"description" bson:"description"`
	Notification string `json:"notification" bson:"notification"`
	Frequency    string `json:"frequency" bson:"frequency"` // (contains time offset in milliseconds)
	Priority     int    `json:"priority" bson:"priority"`
}

type CreateEventOptions struct {
	CalendarId   string `json:"calendarId" bson:"calendar_id"`
	Name         string `json:"name" bson:"name"`
	Start        int64  `json:"start" bson:"start"`
	End          int64  `json:"end" bson:"end"`
	Location     string `json:"location" bson:"location"`
	Description  string `json:"description" bson:"description"`
	Notification string `json:"notification" bson:"notification"`
	Frequency    string `json:"frequency" bson:"frequency"`
	Priority     int    `json:"priority" bson:"priority"`
}

type UpdateEventOptions struct {
	Name         string `json:"name" bson:"name"`
	Start        int64  `json:"start" bson:"start"`
	End          int64  `json:"end" bson:"end"`
	Location     string `json:"location" bson:"location"`
	Description  string `json:"description" bson:"description"`
	Notification string `json:"notification" bson:"notification"`
	Frequency    string `json:"frequency" bson:"frequency"`
	Priority     int    `json:"priority" bson:"priority"`
}

func CreateEvent(c *gin.Context, options *CreateEventOptions) (*Event, error) {
	event := Event{
		EventId:      gonanoid.Must(),
		CalendarId:   options.CalendarId,
		Name:         options.Name,
		Start:        options.Start,
		End:          options.End,
		Location:     options.Location,
		Description:  options.Description,
		Notification: options.Notification,
		Frequency:    options.Frequency,
		Priority:     options.Priority,
	}

	_, err := Events.InsertOne(c, event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func FetchEventById(c *gin.Context, eventId string) (*Event, error) {
	var event Event
	filter := bson.D{{"_id", eventId}}
	result := Events.FindOne(c, filter)
	if err := result.Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func FetchEventsByCalendarId(c *gin.Context, calendarId string) (*[]Event, error) {
	var events []Event
	filter := bson.D{{"calendar_id", calendarId}}
	cursor, err := Events.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &events); err != nil {
		return nil, err
	}

	return &events, nil
}

func UpdateEvent(c *gin.Context, id string, options *UpdateEventOptions) (*Event, error) {
	update := bson.D{
		bson.E{"name", options.Name},
		bson.E{"start", options.Start},
		bson.E{"end", options.End},
		bson.E{"location", options.Location},
		bson.E{"description", options.Description},
		bson.E{"notification", options.Notification},
		bson.E{"frequency", options.Frequency},
		bson.E{"priority", options.Priority},
	}

	opts := bson.D{{"$set", update}}

	updateResult, err := Events.UpdateByID(c, id, opts)
	if err != nil {
		return nil, err
	}

	if updateResult.MatchedCount != 1 {
		return nil, errors.New("event not found")
	}

	var event Event
	filter := bson.D{{"_id", id}}
	findResult := Events.FindOne(c, filter)
	if err = findResult.Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}
