package database

import (
	"context"

	_ "github.com/joho/godotenv/autoload"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// user for sending in api
type UserResponse struct {
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Chatrooms []string `json:"chatrooms"`
	Uid       string   `json:"uid"`
}

// user for getting data in req
type User struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	Username  string   `json:"username"`
	Chatrooms []string `json:"chatrooms"`
	Uid       string   `json:"uid"`
}

// user internal
type user struct {
	Email     string
	Password  string
	Username  string
	Chatrooms []string
	DbId      primitive.ObjectID `bson:"_id"`
}

// convering function
func (u *user) toPublicUser() User {
	user := User{}

	user.Email = u.Email
	user.Username = u.Username
	user.Chatrooms = u.Chatrooms
	user.Uid = u.DbId.Hex()

	return user
}

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

func (s *service) GetUserByEmail(email string) (User, error) {
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

func (s *service) GetUser(uid string) (User, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	_id, err := primitive.ObjectIDFromHex(uid)

	if err != nil {
		return User{}, err
	}

	res := coll.FindOne(context.TODO(), bson.D{
		primitive.E{
			Key: "_id", Value: _id,
		},
	})

	if res.Err() != nil {
		return User{}, res.Err()
	}

	userReceiver := user{}
	err = res.Decode(&userReceiver)

	if err != nil {
		return User{}, err
	}

	user := userReceiver.toPublicUser()

	return user, nil

}

func (s *service) GetUsers() ([]UserResponse, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	cursor, err := coll.Find(context.TODO(), bson.D{})

	if err != nil {
		return []UserResponse{}, err
	}

	usersDb := []user{}
	err = cursor.All(context.TODO(), &usersDb)

	if err != nil {
		return []UserResponse{}, err
	}

	users := []UserResponse{}

	for _, val := range usersDb {
		users = append(users, UserResponse{Username: val.Username, Email: val.Email, Chatrooms: val.Chatrooms, Uid: val.DbId.Hex()})
	}

	return users, nil
}

func (s *service) AddUserChatroom(uidUser, uidChatroom string) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Users")

	id, _ := primitive.ObjectIDFromHex(uidUser)

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "chatrooms", Value: uidChatroom}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}
