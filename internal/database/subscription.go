package database

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Subscription struct {
	SubscriptionId string `json:"subscriptionId" bson:"_id"`
	SubscriberId   string `json:"subscriberId" bson:"subscriber_id"`
	CalendarId     string `json:"calendarId" bson:"calendar_id"`
}

type SubscribeOptions struct {
	SubscriberId string `json:"subscriberId"`
	CalendarId   string `json:"calendarId"`
}

type UnsubscribeOptions struct {
	UserId     string
	CalendarId string
}

func CreateSubscription(c *gin.Context, options *SubscribeOptions) (*Subscription, error) {
	sub := Subscription{
		SubscriptionId: options.SubscriberId + options.CalendarId,
		SubscriberId:   options.SubscriberId,
		CalendarId:     options.CalendarId,
	}

	_, err := Subscriptions.InsertOne(c, sub)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

func DeleteSubscription(c *gin.Context, subscriberId string, calendarId string) error {
	filter := bson.D{
		{"_id", subscriberId + calendarId},
	}

	result, err := Subscriptions.DeleteOne(c, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

func FetchAllSubscriptions(c *gin.Context, subscriberId string) (*[]Subscription, error) {
	var subscriptions []Subscription
	filter := bson.D{{"subscriber_", subscriberId}}
	cursor, err := Subscriptions.Find(c, filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(c, &subscriptions); err != nil {
		return nil, err
	}

	return &subscriptions, nil
}
