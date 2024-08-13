package database

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	AddUser(user User) (User, error)
	GetUser(email string) (User, error)
	UserExists(email string) bool
	CreateChat(creatorUid, targetUid string) (Chatroom, error)
	AddUserChatroom(uidUser, uidChatroom string) error
	WatchChat(chatId string) (*mongo.ChangeStream, error)
	SendMessage(chatId string, message Message) error
	GetChatHistory(chatId string) ([]Message, error)
}

type service struct {
	db *mongo.Client
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
