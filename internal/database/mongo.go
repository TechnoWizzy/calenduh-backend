package database

import (
	"calenduh-backend/internal/util"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var Users *mongo.Collection
var Groups *mongo.Collection
var Events *mongo.Collection
var Sessions *mongo.Collection
var Calendars *mongo.Collection
var GroupMembers *mongo.Collection
var SubscribesTo *mongo.Collection

func New() *mongo.Client {
	var client *mongo.Client
	var err error
	iterations := 0

	for {
		iterations += 1
		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(MakeConnectionUri()))
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if iterations > 30 {
			log.Fatal("Unable to connect to MongoDB after 30 seconds")
		}

		break
	}

	if iterations > 1 {
		fmt.Printf("%d iterations to connect to postgres\n", iterations)
	}

	database := client.Database(util.GetEnv("MONGO_INITDB_DATABASE"))
	Users = database.Collection("users")
	Groups = database.Collection("groups")
	Events = database.Collection("events")
	Sessions = database.Collection("sessions")
	Calendars = database.Collection("calendars")
	GroupMembers = database.Collection("belongs_to")
	SubscribesTo = database.Collection("subscribes_to")

	return client
}

func MakeConnectionUri() string {
	// mongodb://$MONGO_INITDB_ROOT_USERNAME:$MONGO_INITDB_ROOT_PASSWORD@$MONGO_HOST:$MONGO_PORT
	username := util.GetEnv("MONGO_INITDB_ROOT_USERNAME")
	password := util.GetEnv("MONGO_INITDB_ROOT_PASSWORD")
	host := util.GetEnv("MONGO_HOST")
	port := util.GetEnv("MONGO_INTERNAL_PORT")
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, host, port)
}
