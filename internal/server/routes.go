package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go-chat/internal/database"
	authToken "go-chat/internal/jwt"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/health", s.healthHandler)

	r.POST("/auth", s.createUserHandler)

	r.POST("/auth/login", s.loginHandler)

	r.POST("/chat", s.createChatHandler)

	r.GET("/chat", s.connectToChatHandler)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) createUserHandler(c *gin.Context) {
	user := database.User{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	if s.db.UserExists(user.Email) {
		c.JSON(http.StatusConflict, gin.H{"message": "user already exists"})
		return
	}

	h := sha256.New()
	h.Write([]byte(user.Password))
	password := h.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	user, err = s.db.AddUser(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong while adding user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": authToken.CreateToken(user),
	})
}

type WrongPasswordErr struct{}

func (err WrongPasswordErr) Error() string {
	return "passwords don't match"
}

func (s *Server) loginHandler(c *gin.Context) {
	user := database.User{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	h := sha256.New()
	h.Write([]byte(user.Password))
	password := h.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	userDb, err := s.db.GetUser(user.Email)

	if err == nil && user.Password != userDb.Password {
		err = WrongPasswordErr{}
	}

	if errors.Is(err, WrongPasswordErr{}) || errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusUnauthorized, gin.H{"message:": "Wrong email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": authToken.CreateToken(userDb),
	})
}

type createChatPayload struct {
	Jwt        string `json:"jwt"`
	CreatorUid string `json:"creatorUid"`
	TargetUid  string `json:"targetUid"`
}

func (s *Server) createChatHandler(c *gin.Context) {
	payload := createChatPayload{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &payload)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	err = authToken.VerifyToken(payload.Jwt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}

	chatroom, err := s.db.CreateChat(payload.CreatorUid, payload.TargetUid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"chatroomID": chatroom.ChatId})
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type chatPayload struct {
	Jwt    string `json:"jwt"`
	ChatId string `json:"chatId"`
}

func findVal(data map[string]interface{}, val string) string {
	if updateDescription, ok := data["updateDescription"].(map[string]interface{}); ok {
		if updatedFields, ok := updateDescription["updatedFields"].(map[string]interface{}); ok {
			for _, v := range updatedFields {
				if field, ok := v.(map[string]interface{}); ok {
					if value, ok := field[val].(string); ok {
						return value
					}
				}
			}
		}
	}
	return ""
}

func awaitForMsg(chatId string, uid string, conn *ws.Conn, s *Server, closed <-chan bool) {
	stream, _ := s.db.WatchChat(chatId)
	defer stream.Close(context.TODO())

	for {
		select {
		case <-closed:
			return
		default:
			var changeEvent map[string]interface{}
			stream.Next(context.TODO())
			curr := stream.Current
			stream.Decode(&changeEvent)

			json.Unmarshal([]byte(curr.String()), &changeEvent)

			message := findVal(changeEvent, "message")
			sender := findVal(changeEvent, "sender")

			if sender != uid {
				conn.WriteJSON(database.Message{Message: message, Sender: sender})
			}
		}
	}

}

func (s *Server) connectToChatHandler(c *gin.Context) {
	writer := c.Writer
	request := c.Request
	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		return
	}
	defer conn.Close()

	chatData := chatPayload{}
	conn.ReadJSON(&chatData)

	err = authToken.VerifyToken(chatData.Jwt)

	if err != nil {
		conn.WriteJSON(wsError{Code: http.StatusUnauthorized, Message: "bad user data"})
	}

	user, _ := authToken.GetUserData(chatData.Jwt)

	userFromDb, _ := s.db.GetUser(user.Email)
	hasAcces := false
	for _, id := range userFromDb.Chatrooms {
		if chatData.ChatId == id {
			hasAcces = true
		}
	}

	if !hasAcces {
		conn.WriteJSON(wsError{Code: http.StatusForbidden, Message: "you are not able to acces this chat"})
		return
	}

	history, err := s.db.GetChatHistory(chatData.ChatId)

	if err != nil {
		fmt.Println(err)
	}

	conn.WriteJSON(gin.H{"history": history})

	closed := make(chan bool)

	defer func() { closed <- true }()
	go awaitForMsg(chatData.ChatId, user.Uid, conn, s, closed)

	for {
		message := database.Message{}
		err := conn.ReadJSON(&message)

		if err != nil {
			fmt.Println(err)
			return
		}

		err = s.db.SendMessage(chatData.ChatId, message)

		if err != nil {
			fmt.Println(err)
		}
	}
}
