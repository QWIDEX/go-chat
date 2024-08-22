package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	AddUser(user User) (User, error)
	GetUser(email string) (User, error)
	AddUserChatroom(uidUser, uidChatroom string) error
	CreateChatroom(creatorUid, targetUid string) (Chatroom, error)
	AddChatroomMember(uid, chatid string) error
	SendMessage(chatId string, message Message) error
	GetChat(chatId string) (Chatroom, error)
}

type service struct {
	db *mongo.Client
}

var (
	// Uri = os.Getenv("mongoURI")
	username = os.Getenv("DB_USERNAME")
	password = os.Getenv("DB_ROOT_PASSWORD")
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
)

func New() Service {
	//connect using atlas uri
	// client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Uri))

	//connect with docker
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s:%s/", username, password, host, port)))

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
