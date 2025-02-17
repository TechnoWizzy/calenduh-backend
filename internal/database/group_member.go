package database

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type GroupMember struct {
	GroupMemberId string `json:"groupMemberId" bson:"_id"`
	UserId        string `json:"userId" bson:"user_id"`
	GroupId       string `json:"groupId" bson:"group_id"`
}

type CreateGroupMemberOptions struct {
	UserId  string
	GroupId string
}

func CreateGroupMember(c *gin.Context, options *CreateGroupMemberOptions) (*GroupMember, error) {
	group := GroupMember{
		GroupMemberId: options.UserId + "_" + options.GroupId,
		UserId:        options.UserId,
		GroupId:       options.GroupId,
	}

	_, err := GroupMembers.InsertOne(c, group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func FetchGroupMemberByUserId(c *gin.Context, id string) (*GroupMember, error) {
	var user GroupMember
	filter := bson.D{{"userId", id}}
	result := GroupMembers.FindOne(c, filter)
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
