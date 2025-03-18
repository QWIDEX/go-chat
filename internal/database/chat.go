package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	Sender  string             `json:"sender"`
	Message string             `json:"message"`
	SentAt  primitive.DateTime `json:"sentAt" bson:"sentAt"`
}

type Chatroom struct {
	ChatName    string             `json:"chatName"`
	Members     []string           `json:"members"`
	Chat        []Message          `json:"chat"`
	ChatId      string             `json:"chatId"`
	LastMessage primitive.DateTime `json:"lastMessage"`
}

func (s *service) CreateChatroom(creatorUid, targetUid, chatName string) (Chatroom, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	chatroom := Chatroom{}

	res, err := coll.InsertOne(context.TODO(), map[string]any{
		"members":     []string{creatorUid, targetUid},
		"chat":        []Message{},
		"chatName":    chatName,
		"lastMessage": primitive.NewDateTimeFromTime(time.Now()),
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

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "members", Value: uid}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}

func (s *service) SendMessage(chatId string, message Message) error {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	sentAt := primitive.NewDateTimeFromTime(time.Now())

	message.SentAt = sentAt

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "chat", Value: bson.D{{Key: "$each", Value: []Message{message}}, {Key: "$position", Value: 0}}}}}, {Key: "$set", Value: bson.D{{Key: "lastMessage", Value: sentAt}}}}
	_, err := coll.UpdateByID(context.TODO(), id, update)

	return err
}

func (s *service) GetChat(chatId string, skip, length int) (Chatroom, error) {
	db := s.db.Database("ChatApp")
	coll := db.Collection("Chats")

	id, _ := primitive.ObjectIDFromHex(chatId)

	if length <= 0 {
		length = 1
	}

	projection := bson.D{{Key: "chat", Value: bson.D{{Key: "$slice", Value: bson.A{skip, length}}}}}

	opt := options.FindOne().SetProjection(projection)

	res := coll.FindOne(
		context.TODO(),
		bson.D{
			primitive.E{
				Key: "_id", Value: id,
			},
		},
		opt,
	)

	if res.Err() != nil {
		return Chatroom{}, res.Err()
	}

	chatroom := Chatroom{}
	err := res.Decode(&chatroom)

	if err != nil {
		return Chatroom{}, err
	}

	chatroom.ChatId = chatId

	return chatroom, nil
}
