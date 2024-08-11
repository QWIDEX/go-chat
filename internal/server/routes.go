package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"go-chat/internal/database"
	authToken "go-chat/internal/jwt"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/health", s.healthHandler)

	r.POST("/auth", s.createUserHandler)

	r.POST("/auth/login", s.loginHandler)

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

	user, err = s.db.ValidatePassword(user)

	if errors.Is(err, database.WrongPasswordErr{}) || errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusUnauthorized, gin.H{"message:": "Wrong email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": authToken.CreateToken(user),
	})
}
