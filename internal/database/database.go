package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	AddUser(user User) (User, error)
	GetUser(email string) (User, error)
	UserExists(email string) bool
	ValidatePassword(user User) (User, error)
}

type service struct {
	db *mongo.Client
}

var (
	Uri = os.Getenv("mongoURI")
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
	Uid      string `json:"uid"`
}

type user struct {
	Email    string
	Password string
	Username string
	DbId     primitive.ObjectID `bson:"_id"`
}

func (u *user) toPublicUser() User {
	user := User{}

	user.Email = u.Email
	user.Password = u.Password
	user.Username = u.Username
	user.Uid = u.DbId.Hex()

	return user
}

func New() Service {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Uri))

	if err != nil {
		log.Fatal(err)

	}
	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
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

type WrongPasswordErr struct{}

func (err WrongPasswordErr) Error() string {
	return "passwords don't match"
}

func (s *service) ValidatePassword(userReq User) (User, error) {
	user, err := s.GetUser(userReq.Email)

	if err != nil {
		return User{}, err
	}

	if userReq.Password == user.Password {
		return user, nil
	}

	return User{}, WrongPasswordErr{}
}
