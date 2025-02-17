package database

import (
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type Group struct {
	GroupId string `json:"group_id" bson:"_id"`
	Name    string `json:"name" bson:"name"`
}

type CreateGroupOptions struct {
	Name string
}

type UpdateGroupOptions struct {
	Name string
}

func CreateGroup(c *gin.Context, options *CreateGroupOptions) (*Group, error) {
	group := Group{
		GroupId: gonanoid.Must(),
		Name:    options.Name,
	}

	_, err := Groups.InsertOne(c, group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func FetchGroupById(c *gin.Context, id string) (*Group, error) {
	var user Group
	filter := bson.D{{"_id", id}}
	result := Groups.FindOne(c, filter)
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchAllGroups(c *gin.Context) (*[]Group, error) {
	var users []Group
	filter := bson.D{}
	cursor, err := Groups.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &users); err != nil {
		return nil, err
	}

	return &users, nil
}

func UpdateGroup(c *gin.Context, id string, options *UpdateGroupOptions) (*Group, error) {
	opts := bson.D{{"$set", bson.A{
		bson.D{{"name", options.Name}},
	}}}

	_, err := Groups.UpdateByID(c, id, opts)
	if err != nil {
		return nil, err
	}

	var user Group
	filter := bson.D{{"_id", id}}
	result := Groups.FindOne(c, filter)
	if err = result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
