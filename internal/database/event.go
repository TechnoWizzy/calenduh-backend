package database

import (
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Event struct {
	EventId			string 		`json:"event_id" bson:"_id"`
	CalendarId 		string 		`json:"calendar_id" bson:"calendar_id"`
	Name			string 		`json:"name" bson:"name"`	
	Start 			time.Time 	`json:"start" bson:"start"` // Start date and time
	End 			time.Time 	`json:"end" bson:"end"` // End date and time
	Location 		string 		`json:"location" bson:"location"`
	Description 	string 		`json:"description" bson:"description"`
	Notification 	string 		`json:"notification" bson:"notification"` // (contains time offset in milliseconds)
	Frequency 		string 		`json:"frequency" bson:"frequency"`
	Priority 		int 		`json:"priority" bson:"priority"`
}

type CreateEventOptions struct {
	EventId 		string
	CalendarId 		string
	Name 			string
	Start 			time.Time
	End 			time.Time
	Location 		string
	Description 	string
	Notification 	string
	Frequency 		string
	Priority 		int
}

type UpdateEventOptions struct {
	Name 			string
	Start 			time.Time
	End 			time.Time
	Location 		string
	Description 	string
	Notification 	string
	Frequency 		string
	Priority 		int
}

func CreateEvent(c *gin.Context, options *CreateEventOptions) (*Event, error) {
	event := Event {
		EventId:		gonanoid.Must(),
		CalendarId: 	options.CalendarId, // TODO: where is calendarId retrieved from? see calendar.go
		Name: 			options.Name,
		Start: 			options.Start,
		End: 			options.End,
		Location: 		options.Location,
		Description:	options.Description,
		Notification: 	options.Notification,
		Frequency: 		options.Frequency,
		Priority:		options.Priority,
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
	if !options.Start.IsZero() {
		update = append(update, bson.E{"start", options.Start})
	}
	if !options.End.IsZero() {
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
		return nil, nil
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

// GETTERS
func (e *Event) GetCalendarId() string {
	return e.CalendarId
}

func (e *Event) GetEventId() string {
	return e.EventId
}

func (e *Event) GetName() string {
	return e.Name
}

func (e *Event) GetStartDate() time.Time {
	return e.Start
}

func (e *Event) GetEndDate() time.Time {
	return e.End
}

func (e *Event) GetLocation() string {
	return e.Location
}

func (e *Event) GetDescription() string {
	return e.Description
}

func (e *Event) GetNotification() string {
	return e.Notification
}

func (e *Event) GetFrequency() string {
	return e.Frequency
}

func (e *Event) GetPriority() int {
	return e.Priority
}

// SETTERS
func (e *Event) SetCalendarId(calendarId string) {
	e.CalendarId = calendarId
}

func (e *Event) SetEventId(eventId string) {
	e.EventId = eventId
}

func (e *Event) SetName(name string) {
    e.Name = name
}

func (e *Event) SetStartDate(start time.Time) {
    e.Start = start
}

func (e *Event) SetEndDate(end time.Time) {
    e.End = end
}

func (e *Event) SetLocation(location string) {
    e.Location = location
}

func (e *Event) SetDescription(description string) {
    e.Description = description
}

func (e *Event) SetNotification(notification string) {
    e.Notification = notification
}

func (e *Event) SetFrequency(frequency string) {
    e.Frequency = frequency
}

func (e *Event) SetPriority(priority int) {
    e.Priority = priority
}