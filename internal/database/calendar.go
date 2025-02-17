package database

import (
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Calendar struct {
	CalendarId string `json:"calendarId" bson:"_id"`
	Type       string `json:"type" bson:"type"`
	Title      string `json:"title" bson:"title"`
	OwnerId    string `json:"ownerId" bson:"owner_id"`
	IsPublic   bool   `json:"isPublic" bson:"is_public"`
}

type CreateCalendarOptions struct {
	Type     string
	Title    string
	OwnerId  string
	IsPublic bool
}

type UpdateCalendarOptions struct {
	Title    string
	IsPublic bool
}

const (
	UserCalendar  = "USER"
	GroupCalendar = "GROUP"
)

func CreateCalendar(c *gin.Context, options *CreateCalendarOptions) (*Calendar, error) {
	calendar := Calendar{
		CalendarId: gonanoid.Must(),
		Type:       options.Type,
		Title:      options.Title,
		OwnerId:    options.OwnerId,
		IsPublic:   false,
	}

	_, err := Calendars.InsertOne(c, calendar)
	if err != nil {
		return nil, err
	}

	return &calendar, nil
}

func FetchCalendarById(c *gin.Context, calendarId string) (*Calendar, error) {
	var calendar Calendar
	filter := bson.D{{"_id", calendarId}}
	result := Calendars.FindOne(c, filter)
	if err := result.Decode(&calendar); err != nil {
		return nil, err
	}

	return &calendar, nil
}

func FetchCalendarsByOwnerId(c *gin.Context, ownerId string) (*[]Calendar, error) {
	var calendars []Calendar
	filter := bson.D{{"owner_id", ownerId}}
	cursor, err := Calendars.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &calendars); err != nil {
		return nil, err
	}

	return &calendars, nil
}

func UpdateCalendar(c *gin.Context, calendarId string, options *UpdateCalendarOptions) (*Calendar, error) {
	opts := bson.D{{"$set", bson.A{
		bson.D{{"title", options.Title}},
		bson.D{{"is_public", options.IsPublic}},
	}}}

	_, err := Calendars.UpdateByID(c, calendarId, opts)
	if err != nil {
		return nil, err
	}

	var user Calendar
	filter := bson.D{{"_id", calendarId}}
	result := Calendars.FindOne(c, filter)
	if err = result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
