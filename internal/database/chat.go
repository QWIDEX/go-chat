package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *service) WatchChat(chatId string) (*mongo.ChangeStream, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	stream, err := coll.Watch(context.TODO(), mongo.Pipeline{
		{{"$match", bson.D{
			{"documentKey._id", id},
			{"operationType", "update"},
		}}},
		bson.D{{"$project", bson.D{{"updateDescription.updatedFields", true}}}},
	},
	)

	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *service) SendMessage(chatId string, message Message) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	update := bson.D{{"$push", bson.D{{"chat", message}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}

func (s *service) GetChatHistory(chatId string) ([]Message, error) {
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
		return []Message{}, err
	}

	return chatroom.Chat, nil
}
