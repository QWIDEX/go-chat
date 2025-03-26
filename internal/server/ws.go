package server

import (
	"encoding/json"
	"fmt"
	"go-chat/internal/database"
	authToken "go-chat/internal/jwt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// type of error that will be sent through websocket
type serverMessage struct {
	Code    int    `json:"code"`
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

type wsMessage struct {
	Message string `json:"message"`
	Dest    string `json:"dest"`
}

type auth struct {
	AccessToken string `json:"accessToken"`
}

func (s *Server) connectToWS(c *gin.Context) {
	// creating websocket connection
	writer := c.Writer
	request := c.Request
	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// authentificating a user
	_, msg, err := conn.ReadMessage()

	if err != nil {
		conn.WriteJSON(serverMessage{Code: http.StatusBadRequest, Message: "unable to read payload", Sender: "server"})
		fmt.Println(err)
		return
	}

	auth := auth{}
	err = json.Unmarshal(msg, &auth)
	if err != nil {
		conn.WriteJSON(serverMessage{Code: http.StatusBadRequest, Message: "unable to read json", Sender: "server"})
		fmt.Println(err)
		return
	}

	jwt := auth.AccessToken

	if len(jwt) == 0 {
		conn.WriteJSON(serverMessage{Code: http.StatusUnauthorized, Message: "failed to authentificate", Sender: "server"})
		return
	}

	err = authToken.VerifyToken(jwt)

	if err != nil {
		fmt.Println(err)
		conn.WriteJSON(serverMessage{Code: http.StatusUnauthorized, Message: "failed to authentificate", Sender: "server"})
		return
	}

	user, err := authToken.GetUserData(jwt)

	if err != nil {
		fmt.Println(err)
		conn.WriteJSON(serverMessage{Code: http.StatusInternalServerError, Message: "something went wrong", Sender: "server"})
		return
	}

	// adding connection to a list of server websocket connections
	s.wsConnoctions[user.Uid] = conn
	defer delete(s.wsConnoctions, user.Uid)

	activeChats := map[string]database.Chatroom{}

	for {
		// awaiting for message
		_, msg, err := conn.ReadMessage()

		if err != nil {
			conn.WriteJSON(serverMessage{Code: http.StatusBadRequest, Message: "unable to read payload", Sender: "server"})
			fmt.Println(err)
			return
		}

		message := wsMessage{}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			fmt.Println(msg)
			conn.WriteJSON(serverMessage{Code: http.StatusBadRequest, Message: "unable to read json", Sender: "server"})
			fmt.Println(err)
			return
		}

		// getting chat data and checking
		if _, ok := activeChats[message.Dest]; !ok {
			chat, err := s.db.GetChat(message.Dest, 0, 1)

			if err != nil {
				fmt.Println(err)
				conn.WriteJSON(serverMessage{Code: http.StatusInternalServerError, Message: "something went wrong", Sender: "server"})
				return
			}

			// validating acces

			isMember := false
			for _, uid := range chat.Members {
				if uid == user.Uid {
					isMember = true
					break
				}
			}

			if !isMember {
				conn.WriteJSON(serverMessage{Code: http.StatusForbidden, Message: "Access denied", Sender: "server"})
				return
			}

			activeChats[message.Dest] = chat
			defer delete(activeChats, message.Dest)
		}

		chat := activeChats[message.Dest]

		// sending message to a db
		err = s.db.SendMessage(message.Dest, database.Message{Message: message.Message, Sender: user.Username})

		if err != nil {
			fmt.Println(err)
			conn.WriteJSON(serverMessage{Code: http.StatusInternalServerError, Message: "something went wrong", Sender: "server"})
		}

		// sending message to other websocket clients that have acces to this chat

		for _, uid := range chat.Members {
			if uid != user.Uid {
				otherConn, ok := s.wsConnoctions[uid]
				if ok {
					otherConn.WriteJSON(gin.H{"sender": user.Username, "message": message.Message, "sentAt": primitive.NewDateTimeFromTime(time.Now()), "dest": message.Dest})
				}
			} else {
				conn.WriteJSON(serverMessage{Code: http.StatusOK, Message: "delivered", Sender: "server"})
			}
		}

	}
}
