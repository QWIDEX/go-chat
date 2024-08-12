package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Sender  string
	Message string
}

type Chatroom struct {
	Members []string
	Chat    []Message
	ChatId  string
}

func (s *service) CreateChat(creatorUid, targetUid string) (Chatroom, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	chatroom := Chatroom{}

	res, err := coll.InsertOne(context.TODO(), map[string]any{
		"creator": creatorUid,
		"target":  targetUid,
		"chat":    []Message{},
	})

	if err != nil {
		return Chatroom{}, err
	}

	id, _ := res.InsertedID.(primitive.ObjectID)

	chatroom.ChatId = id.Hex()

	err = s.AddUserChatroom(creatorUid, chatroom.ChatId)

	if err != nil {
		return Chatroom{}, err
	}

	err = s.AddUserChatroom(targetUid, chatroom.ChatId)

	if err != nil {
		return Chatroom{}, err
	}

	chatroom.Members = append(chatroom.Members, creatorUid, targetUid)

	return chatroom, nil
}
