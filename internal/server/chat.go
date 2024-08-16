package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-chat/internal/database"
	authToken "go-chat/internal/jwt"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

// payload that is expected to create chat
type createChatPayload struct {
	Jwt        string `json:"jwt"`
	CreatorUid string `json:"creatorUid"`
	TargetUid  string `json:"targetUid"`
}

func (s *Server) createChat(creatorUid, targetUid string) (database.Chatroom, error) {
	// creating chatroom
	chatroom, err := s.db.CreateChatroom(creatorUid, targetUid)

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
	// getting payload
	payload := createChatPayload{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}
	// verifying token
	err = authToken.VerifyToken(payload.Jwt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}
	// creating chatroom
	chatroom, err := s.createChat(payload.CreatorUid, payload.TargetUid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"chatroomID": chatroom.ChatId})
}

// webocket upgrader config
var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// type of error that will be sent through websocket
type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// payload that is expected to be user's first message
type chatPayload struct {
	Jwt    string `json:"jwt"`
	ChatId string `json:"chatId"`
}

func (s *Server) connectToChatHandler(c *gin.Context) {
	// creating websocket connection
	writer := c.Writer
	request := c.Request
	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return
	}
	defer conn.Close()

	//getting data

	// getting payload
	chatData := chatPayload{}
	conn.ReadJSON(&chatData)

	// verifying token
	err = authToken.VerifyToken(chatData.Jwt)

	if err != nil {
		conn.WriteJSON(wsError{Code: http.StatusUnauthorized, Message: "token not valid"})
	}
	// getting basic user data
	user, _ := authToken.GetUserData(chatData.Jwt)

	// adding connection to a list of server websocket connections

	s.wsConnoctions[user.Uid] = conn
	defer delete(s.wsConnoctions, user.Uid)

	// getting chatroom data

	chat, err := s.db.GetChat(chatData.ChatId)

	// checking if user has acces to chatroom
	hasAcces := false
	for _, id := range chat.Members {
		if user.Uid == id {
			hasAcces = true
		}
	}

	if !hasAcces {
		conn.WriteJSON(wsError{Code: http.StatusForbidden, Message: "you are not able to acces this chat"})
		return
	}

	if err != nil {
		fmt.Println(err)
	}
	// senfing chatroom history to user
	conn.WriteJSON(gin.H{"history": chat.Chat})

	for {
		// awaiting for message
		message := database.Message{}
		err := conn.ReadJSON(&message)

		if err != nil {
			fmt.Println(err)
			return
		}
		// sending message to a db
		err = s.db.SendMessage(chatData.ChatId, message)
		// sending message to other websocket clients that have acces to this chat
		if err == nil {
			for _, uid := range chat.Members {
				if uid != user.Uid {
					otherConn, ok := s.wsConnoctions[uid]
					if ok {
						otherConn.WriteJSON(message)
					}
				}
			}
		}

		if err != nil {
			fmt.Println(err)
		}
	}
}

type addChatroomMemberPayload struct {
	Jwt       string `json:"jwt"`
	ChatId    string `json:"chatId"`
	MemberUid string `json:"memberUid"`
}

func (s *Server) addChatroomMemberHandler(c *gin.Context) {
	payload := addChatroomMemberPayload{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	// verifying jwt

	err = authToken.VerifyToken(payload.Jwt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "bad user data"})
		return
	}

	// getting data

	userData, _ := authToken.GetUserData(payload.Jwt)

	chat, err := s.db.GetChat(payload.ChatId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	// checking if user has acces to chatroom
	hasAcces := false
	for _, id := range chat.Members {
		if userData.Uid == id {
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
