package database

import (
	"context"
	"errors"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Uri = os.Getenv("mongoURI")
)

type User struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Username  string   `json:"username"`
	Chatrooms []string `json:"chatrooms"`
	Uid       string   `json:"uid"`
}

type user struct {
	Email     string
	Password  string
	Username  string
	Chatrooms []string
	DbId      primitive.ObjectID `bson:"_id"`
}

func (u *user) toPublicUser() User {
	user := User{}

	user.Email = u.Email
	user.Password = u.Password
	user.Username = u.Username
	user.Chatrooms = u.Chatrooms
	user.Uid = u.DbId.Hex()

	return user
}

// func (u *User) toPrivateUser() user {
// 	user := user{}

// 	user.Email = u.Email
// 	user.Password = u.Password
// 	user.Username = u.Username
// 	user.DbId, _ = primitive.ObjectIDFromHex(u.Uid)

// 	return user
// }

func (s *service) AddUser(user User) (User, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	res, err := coll.InsertOne(context.TODO(), map[string]any{
		"email":    user.Email,
		"username": user.Username,
		"password": user.Password,
	})

	if err != nil {
		return User{}, err
	}

	id, ok := res.InsertedID.(primitive.ObjectID)

	if ok {
		user.Uid = id.Hex()
	}

	return user, nil
}

func (s *service) GetUser(email string) (User, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	res := coll.FindOne(context.TODO(), bson.D{
		primitive.E{
			Key: "email", Value: email,
		},
	})

	if res.Err() == nil {
		userReceiver := user{}
		err := res.Decode(&userReceiver)

		user := userReceiver.toPublicUser()

		if err != nil {
			return User{}, err
		}

		return user, nil
	}

	return User{}, res.Err()
}

func (s *service) UserExists(email string) bool {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	if res := coll.FindOne(context.TODO(), bson.D{
		primitive.E{
			Key: "email", Value: email,
		},
	}); errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return false
	}

	return true
}

func (s *service) AddUserChatroom(uidUser, uidChatroom string) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	id, _ := primitive.ObjectIDFromHex(uidUser)

	update := bson.D{{"$push", bson.D{{"chatrooms", uidChatroom}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}
