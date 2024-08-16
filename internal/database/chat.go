package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

type Chatroom struct {
	Members []string
	Chat    []Message
	ChatId  string
}

func (s *service) CreateChatroom(creatorUid, targetUid string) (Chatroom, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	chatroom := Chatroom{}

	res, err := coll.InsertOne(context.TODO(), map[string]any{
		"members": []string{creatorUid, targetUid},
		"chat":    []Message{},
	})

	if err != nil {
		return Chatroom{}, err
	}

	id, _ := res.InsertedID.(primitive.ObjectID)

	chatroom.ChatId = id.Hex()

	return chatroom, nil
}

func (s *service) AddChatroomMember(uid, chatid string) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatid)

	update := bson.D{{"$push", bson.D{{"members", uid}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}

func (s *service) SendMessage(chatId string, message Message) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	update := bson.D{{"$push", bson.D{{"chat", message}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}

func (s *service) GetChat(chatId string) (Chatroom, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	res := coll.FindOne(context.TODO(), bson.D{
		primitive.E{
			Key: "_id", Value: id,
		},
	})

	chatroom := Chatroom{}
	err := res.Decode(&chatroom)

	if err != nil {
		return Chatroom{}, err
	}

	return chatroom, nil
}
