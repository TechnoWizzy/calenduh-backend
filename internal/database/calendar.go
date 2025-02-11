package database

import (
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Calendar struct {
	Id       string `json:"id" bson:"_id"`
	Type     string `json:"type" bson:"type"`
	Title    string `json:"title" bson:"title"`
	OwnerId  string `json:"owner_id" bson:"owner_id"`
	IsPublic bool   `json:"is_public" bson:"is_public"`
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
		Id:       gonanoid.Must(),
		Type:     options.Type,
		Title:    options.Title,
		OwnerId:  options.OwnerId,
		IsPublic: false,
	}

	_, err := Calendars.InsertOne(c, calendar)
	if err != nil {
		return nil, err
	}

	return &calendar, nil
}

func FetchCalendarById(c *gin.Context, id string) (*Calendar, error) {
	var user Calendar
	filter := bson.D{{"_id", id}}
	result := Calendars.FindOne(c, filter)
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchAllCalendars(c *gin.Context) (*[]Calendar, error) {
	var users []Calendar
	filter := bson.D{}
	cursor, err := Calendars.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &users); err != nil {
		return nil, err
	}

	return &users, nil
}

func UpdateCalendar(c *gin.Context, id string, options *UpdateCalendarOptions) (*Calendar, error) {
	opts := bson.D{{"$set", bson.A{
		bson.D{{"title", options.Title}},
		bson.D{{"is_public", options.IsPublic}},
	}}}

	_, err := Calendars.UpdateByID(c, id, opts)
	if err != nil {
		return nil, err
	}

	var user Calendar
	filter := bson.D{{"_id", id}}
	result := Calendars.FindOne(c, filter)
	if err = result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
