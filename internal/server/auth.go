package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go-chat/internal/database"
	authToken "go-chat/internal/jwt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Server) createUserHandler(c *gin.Context) {
	// getting user data
	user := database.User{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	// attempting to get user by this email
	_, err = s.db.GetUser(user.Email)

	if !errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusConflict, gin.H{"message": "user already exists"})
		return
	}

	// hashing user's password
	h := sha256.New()
	h.Write([]byte(user.Password))
	password := h.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	// adding user to db
	user, err = s.db.AddUser(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": authToken.CreateToken(user),
		"user": gin.H{
			"Username": user.Username,
			"uid":      user.Uid,
			"email":    user.Email,
		},
	})
}

type WrongPasswordErr struct{}

func (err WrongPasswordErr) Error() string {
	return "passwords don't match"
}

func (s *Server) loginHandler(c *gin.Context) {
	// getting user data
	user := database.User{}
	body, _ := c.GetRawData()

	err := json.Unmarshal(body, &user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to read JSON payload"})
		return
	}

	// hashing user's password
	h := sha256.New()
	h.Write([]byte(user.Password))
	password := h.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	userDb, err := s.db.GetUserByEmail(user.Email)

	// checking if hashes match
	if err == nil && user.Password != userDb.Password {
		err = WrongPasswordErr{}
	}

	if errors.Is(err, WrongPasswordErr{}) || errors.Is(err, mongo.ErrNoDocuments) {
		c.JSON(http.StatusUnauthorized, gin.H{"message:": "Wrong email or password"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jwt": authToken.CreateToken(userDb),
		"user": gin.H{
			"Username": userDb.Username,
			"uid":      userDb.Uid,
			"email":    userDb.Email,
		},
	})
}

func (s *Server) getUserData(c *gin.Context) {
	userId := c.Param("uid")

	userData, err := s.db.GetUser(userId)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"Username": userData.Username,
			"uid":      userData.Uid,
			"email":    userData.Email,
		},
	})
}

func (s *Server) getUsers(c *gin.Context) {
	users, err := s.db.GetUsers()

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, users)
}
