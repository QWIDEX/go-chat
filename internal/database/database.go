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
	AddUserChatroom(uidUser, uidChatroom string) error
	CreateChatroom(creatorUid, targetUid string) (Chatroom, error)
	SendMessage(chatId string, message Message) error
	GetChat(chatId string) (Chatroom, error)
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
