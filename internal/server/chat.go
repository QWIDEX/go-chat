package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"go-chat/internal/database"
	"go-chat/internal/helpers"

	"github.com/gin-gonic/gin"
)

// payload that is expected to create chat
type createChatPayload struct {
	ChatName  string `json:"chatName"`
	TargetUid string `json:"targetUid"`
}

func (s *Server) createChat(creatorUid, targetUid, chatName string) (database.Chatroom, error) {
	// creating chatroom
	chatroom, err := s.db.CreateChatroom(creatorUid, targetUid, chatName)

	if err != nil {
		return database.Chatroom{}, err
	}
	// adding reference to chatroom to creator
	err = s.db.AddUserChatroom(creatorUid, chatroom.ChatId)

	if err != nil {
		return database.Chatroom{}, err
	}
	// adding reference to chatroom to target
	err = s.db.AddUserChatroom(targetUid, chatroom.ChatId)

	if err != nil {
		return database.Chatroom{}, err
	}

	chatroom.Members = append(chatroom.Members, creatorUid, targetUid)

	return chatroom, nil
}

func (s *Server) createChatHandler(c *gin.Context) {
	_, user, ok := helpers.AuthHelper(c)

	if !ok {
		return
	}
	// getting payload
	payload := createChatPayload{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	// creating chatroom
	chatroom, err := s.createChat(user.Uid, payload.TargetUid, payload.ChatName)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"chatId": chatroom.ChatId})
}

func (s *Server) getChatHistory(c *gin.Context) {
	_, user, ok := helpers.AuthHelper(c)

	if !ok {
		return
	}

	chatId := c.Param("chatId")

	from, err := strconv.Atoi(c.Query("from"))

	if err != nil {
		from = 0
	}

	length, err := strconv.Atoi(c.Query("length"))

	if err != nil {
		length = 0
	}

	chatroom, err := s.db.GetChat(chatId, from, length)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	isMember := false

	for _, userId := range chatroom.Members {
		if userId == user.Uid {
			isMember = true
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"message": "you are not able to acces this chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chatroom": chatroom})

}

type addChatroomMemberPayload struct {
	ChatId    string `json:"chatId"`
	MemberUid string `json:"memberUid"`
}

func (s *Server) addChatroomMemberHandler(c *gin.Context) {
	_, user, ok := helpers.AuthHelper(c)

	if !ok {
		return
	}

	payload := addChatroomMemberPayload{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	chat, err := s.db.GetChat(payload.ChatId, 0, 0)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	// checking if user has acces to chatroom
	hasAcces := false
	for _, id := range chat.Members {
		if user.Uid == id {
			hasAcces = true
		}
	}

	if !hasAcces {
		c.JSON(http.StatusForbidden, gin.H{"message": "you are not able to acces this chat"})
		return
	}

	isAlreadyInChatroom := false

	for _, id := range chat.Members {
		if payload.MemberUid == id {
			isAlreadyInChatroom = true
		}
	}

	if isAlreadyInChatroom {
		c.JSON(http.StatusBadRequest, gin.H{"message": "member is already in chatroom"})
		return
	}

	// adding user to chatroom

	err = s.db.AddChatroomMember(payload.MemberUid, payload.ChatId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	err = s.db.AddUserChatroom(payload.MemberUid, payload.ChatId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	chat.Members = append(chat.Members, payload.MemberUid)

	c.JSON(http.StatusOK, gin.H{"members": chat.Members})
}

func (s *Server) getChats(c *gin.Context) {
	_, userJwt, ok := helpers.AuthHelper(c)

	if !ok {
		return
	}

	user, err := s.db.GetUser(userJwt.Uid)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	chatData := []database.Chatroom{}

	for _, val := range user.Chatrooms {
		chatroom, err := s.db.GetChat(val, 0, 20)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
			return
		}
		chatData = append(chatData, chatroom)
	}

	sort.Sort(ByDate(chatData))

	c.JSON(http.StatusOK, gin.H{"chats": chatData})
}

type ByDate []database.Chatroom

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].LastMessage.Time().After(a[j].LastMessage.Time()) }
