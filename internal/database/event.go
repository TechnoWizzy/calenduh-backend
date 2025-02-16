package database

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Event struct {
	Id           string `json:"id" bson:"_id"`
	CalendarId   string `json:"calendar_id" bson:"calendar_id"`
	Name         string `json:"name" bson:"name"`
	Start        int64  `json:"start" bson:"start"` // Start date and time
	End          int64  `json:"end" bson:"end"`     // End date and time
	Location     string `json:"location" bson:"location"`
	Description  string `json:"description" bson:"description"`
	Notification string `json:"notification" bson:"notification"` // (contains time offset in milliseconds)
	Frequency    string `json:"frequency" bson:"frequency"`
	Priority     int    `json:"priority" bson:"priority"`
}

type CreateEventOptions struct {
	CalendarId   string `json:"calendar_id" bson:"calendar_id"`
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
		Id:           gonanoid.Must(),
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

func FetchEventById(c *gin.Context, id string) (*Event, error) {
	var event Event
	filter := bson.D{{"_id", id}}
	result := Events.FindOne(c, filter)
	if err := result.Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}

func FetchAllEvents(c *gin.Context) (*[]Event, error) {
	var events []Event
	filter := bson.D{}
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
	update := bson.D{}
	if options.Name != "" {
		update = append(update, bson.E{"name", options.Name})
	}
	if options.Start != 0 {
		update = append(update, bson.E{"start", options.Start})
	}
	if options.End != 0 {
		update = append(update, bson.E{"end", options.End})
	}
	if options.Location != "" {
		update = append(update, bson.E{"location", options.Location})
	}
	if options.Description != "" {
		update = append(update, bson.E{"description", options.Description})
	}
	if options.Notification != "" {
		update = append(update, bson.E{"notification", options.Notification})
	}
	if options.Frequency != "" {
		update = append(update, bson.E{"frequency", options.Frequency})
	}
	if options.Priority != 0 {
		update = append(update, bson.E{"priority", options.Priority})
	}

	// no fields = do not update anything
	if len(update) == 0 {
		return nil, errors.New("no updates were provided")
	}

	opts := bson.D{{"$set", update}}

	_, err := Events.UpdateByID(c, id, opts)
	if err != nil {
		return nil, err
	}

	var event Event
	filter := bson.D{{"_id", id}}
	result := Events.FindOne(c, filter)
	if err = result.Decode(&event); err != nil {
		return nil, err
	}

	return &event, nil
}

func FetchEventNameById(c *gin.Context, id string) (*Event, error) {
	var event Event
	filter := bson.D{{"_id", id}}
	result := Events.FindOne(c, filter)
	if err := result.Decode(&event); err != nil {
		return nil, err
	}
	return &event, nil
}
