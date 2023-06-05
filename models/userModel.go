package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct { //this struct(which is called in model) act as a middle layer b/w your golang prgm & your db (mongo or any other )
	//db understands json & golang doesn't -> thats why we need a middle layer

	ID            primitive.ObjectID `bson:"_id"`
	First_name    *string            `json: "first_name"  validate:"required,min=2,max=100"`
	Last_name     *string            `json: "last_name"  validate:"required,min=2,max=100"`
	Password      *string            `json: "password" validate:"required,min=6"`
	Email         *string            `json:"email" validate:"email,required"`
	Phone         *string            `json:"phone" validate:"required"`
	Token         *string            `json:"token"`
	User_type     *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER"` //thats similar to the enum concept from js -> there can be 2 kinds of person who can log in
	Refresh_token *string            `json:"refresh_token`
	Created_at    time.Time          `json:"created_at" `
	Updated_at    time.Time          `json:"updated_at" `
	User_id       string             `json:"user_id"`
}
