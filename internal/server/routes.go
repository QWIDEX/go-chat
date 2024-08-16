package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/health", s.healthHandler)

	r.POST("/auth/register", s.createUserHandler)

	r.POST("/auth/login", s.loginHandler)

	r.GET("/user", s.getUserData)

	r.POST("/chat", s.createChatHandler)

	r.PATCH("/chat", s.addChatroomMemberHandler)

	r.GET("/chat", s.connectToChatHandler)

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
