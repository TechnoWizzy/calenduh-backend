package database

import (
	"calenduh-backend/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	Id       string `bson:"_id"`
	Email    string `bson:"email"`
	Username string `bson:"username"`
	Password string `bson:"password"`
}

type CreateUserOptions struct {
	Email    string
	Username string
	Password string
}

type UpdateUserOptions struct {
	Email    string
	Username string
	Password string
}

func CreateUser(c *gin.Context, options *CreateUserOptions) (*User, error) {
	id := gonanoid.Must()
	user := User{
		Id:       id,
		Email:    options.Email,
		Username: options.Username,
		Password: options.Password,
	}

	_, err := Users.InsertOne(c, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchUserById(c *gin.Context, id string) (*User, error) {
	var user User
	filter := bson.D{{"_id", id}}
	result := Users.FindOne(c, filter)
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchUserByEmail(c *gin.Context, email string) (*User, error) {
	var user User
	filter := bson.D{{"email", email}}
	result := Users.FindOne(c, filter)
	if err := result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchAllUsers(c *gin.Context) (*[]User, error) {
	var users []User
	filter := bson.D{}
	cursor, err := Users.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &users); err != nil {
		return nil, err
	}

	return &users, nil
}

func UpdateUser(c *gin.Context, id string, options *UpdateUserOptions) (*User, error) {
	opts := bson.D{{"$set", bson.A{
		bson.D{{"email", options.Email}},
		bson.D{{"username", options.Username}},
		bson.D{{"password", util.GetHash(options.Password + util.GetHash(options.Email))}},
	}}}

	_, err := Users.UpdateByID(c, id, opts)
	if err != nil {
		return nil, err
	}

	var user User
	filter := bson.D{{"_id", id}}
	result := Users.FindOne(c, filter)
	if err = result.Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
