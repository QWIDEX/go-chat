package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/ws", s.connectToWS)

	r.GET("/health", s.healthHandler)

	r.POST("/auth/register", s.createUserHandler)

	r.POST("/auth/login", s.loginHandler)

	// TODO add token invalidation

	// r.POST("/auth/logout")

	r.GET("/auth/refresh-token", s.refreshToken)

	r.GET("/users", s.getUsers)

	r.GET("/users/:uid", s.getUserData)

	r.GET("/chats", s.getChats)

	r.POST("/chats", s.createChatHandler)

	r.PATCH("/chats", s.addChatroomMemberHandler)

	r.GET("/chats/:chatId", s.getChatHistory)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
