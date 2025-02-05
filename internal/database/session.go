package database

import (
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Session struct {
	Id           string `bson:"_id"`
	UserId       string `bson:"user_id"`
	Type         string `bson:"type"`
	AccessToken  string `bson:"access_token"`
	RefreshToken string `bson:"refresh_token"`
	ExpiresOn    int64  `bson:"expires_on"`
}

type CreateSessionOptions struct {
	UserId       string
	Type         string
	AccessToken  string
	RefreshToken string
	ExpiresOn    time.Time
}

const (
	LocalSession = "LOCAL"
)

func CreateSession(c *gin.Context, options *CreateSessionOptions) (*Session, error) {
	session := Session{
		Id:           gonanoid.Must(),
		UserId:       options.UserId,
		Type:         options.Type,
		AccessToken:  options.AccessToken,
		RefreshToken: options.RefreshToken,
		ExpiresOn:    options.ExpiresOn.UnixMilli(),
	}

	c.Get("")

	if _, err := Sessions.InsertOne(c, session); err != nil {
		return nil, err
	}

	return &session, nil
}

func FetchSessionById(c *gin.Context, id string) (*Session, error) {
	filter := bson.D{{"_id", id}}
	result := Sessions.FindOne(c, filter)
	raw, err := result.Raw()
	if err != nil {
		return nil, err
	}

	var session Session
	if err = bson.Unmarshal(raw, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func DeleteSession(c *gin.Context, id string) error {
	filter := bson.D{{"_id", id}}
	_, err := Sessions.DeleteOne(c, filter)
	if err != nil {
		return err
	}

	return nil
}
