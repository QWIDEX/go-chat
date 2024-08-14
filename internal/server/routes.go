package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
