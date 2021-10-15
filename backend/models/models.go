package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Create Struct
type Blog struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Title    string             `json:"title" bson:"title,omitempty"`
	User     *User              `json:"user" bson:"user,omitempty"`
	Tags     []string           `bson:"tags,omitempty"`
	Created  time.Time          `json:"created" bson:"created"`
	Duration int32              `json:"duration" bson:"duration,omitempty"`
}

type User struct {
	//ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Name     string `json:"name,omitempty" bson:"name,omitempty"`
	Email    string `json:"email,omitempty" bson:"email,"`
	Password string `json:"password,omitempty" bson:"password"`
}
