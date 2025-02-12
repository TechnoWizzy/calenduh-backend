package database

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Subscription struct {
	UserId			string `json:"user_id" bson:"user_id"`
	CalendarId 		string `json:"calendar_id" bson:"calendar_id"`
}

type SubscribeOptions struct {
	UserId 		string
	CalendarId 	string
}

type UnsubscribeOptions struct {
	UserId			string
	CalendarId 		string
}

func Subscribe(c *gin.Context, options *SubscribeOptions) (*Subscription, error) {
	sub := Subscription {
		UserId:		options.UserId,
		CalendarId:	options.CalendarId,
	}

	_, err := Subscriptions.InsertOne(c, sub)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}


func Unsubscribe(c *gin.Context, options *SubscribeOptions) (*Subscription, error) {
	sub := Subscription {
		UserId:		options.UserId,
		CalendarId:	options.CalendarId,
	}

	filter := bson.D{
		{"user_id", options.UserId},
		{"calendar_id", options.CalendarId},
	}
	_, err := Subscriptions.DeleteOne(c, filter)
	//_, err := Subscriptions.DeleteOne(c, sub)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

func FetchSubscribedCalendarById(c *gin.Context, id string) (*Subscription, error) {
	var sub Subscription
	filter := bson.D{{"calendar_id", id}}
	result := Subscriptions.FindOne(c, filter)
	if err := result.Decode(&sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func FetchAllSubscriptions(c *gin.Context) (*[]Subscription, error) { // TODO: see if I should return a pointer
	var subs []Subscription
	filter := bson.D{}
	cursor, err := Subscriptions.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &subs); err != nil {
		return nil, err
	}
	return &subs, nil
}

// GETTERS
func (e *SubscribeOptions) GetCalendarId_() string {
	return e.CalendarId
}

func (e *Subscription) GetUserId_() string {
	return e.UserId
}