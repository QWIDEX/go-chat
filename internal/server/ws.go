package server

import (
	"fmt"
	"go-chat/internal/database"
	"go-chat/internal/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
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
	Sender  string `json:"sender"`
	Message string `json:"message"`
	Dest    string `json:"dest"`
}

func (s *Server) connectToWS(c *gin.Context) {
	_, user, ok := helpers.AuthHelper(c)

	if !ok {
		return
	}
	// creating websocket connection
	writer := c.Writer
	request := c.Request
	conn, err := upgrader.Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	// adding connection to a list of server websocket connections

	s.wsConnoctions[user.Uid] = conn
	defer delete(s.wsConnoctions, user.Uid)

	activeChats := map[string]database.Chatroom{}

	for {
		// awaiting for message
		message := wsMessage{}
		err := conn.ReadJSON(&message)

		if err != nil {
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
		err = s.db.SendMessage(message.Dest, database.Message{Message: message.Message, Sender: message.Sender})

		if err != nil {
			fmt.Println(err)
			conn.WriteJSON(serverMessage{Code: http.StatusInternalServerError, Message: "something went wrong", Sender: "server"})
		}

		// sending message to other websocket clients that have acces to this chat

		for _, uid := range chat.Members {
			if uid != user.Uid {
				otherConn, ok := s.wsConnoctions[uid]
				if ok {
					otherConn.WriteJSON(message)
				}
			} else {
				conn.WriteJSON(serverMessage{Code: http.StatusOK, Message: "delivered", Sender: "server"})
			}
		}

	}
}
