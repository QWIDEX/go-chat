package server

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	user, err = s.db.GetUserByEmail(user.Email)

	if user.Uid != "" {
		c.JSON(http.StatusConflict, gin.H{"message": "user already exists"})
		return
	}

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
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

	c.SetCookie(
		"refreshToken",
		authToken.CreateRefreshToken(user),
		int(time.Until(time.Now().AddDate(0, 1, 0)).Seconds()),
		"/",
		"localhost",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken": authToken.CreateAccessToken(user),
		"user": gin.H{
			"Username":  user.Username,
			"uid":       user.Uid,
			"email":     user.Email,
			"chatrooms": user.Chatrooms,
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

	c.SetCookie(
		"refreshToken",
		authToken.CreateRefreshToken(userDb),
		int(time.Until(time.Now().AddDate(0, 1, 0)).Seconds()),
		"/",
		"localhost",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken": authToken.CreateAccessToken(userDb),
		"user": gin.H{
			"Username":  userDb.Username,
			"uid":       userDb.Uid,
			"email":     userDb.Email,
			"chatrooms": userDb.Chatrooms,
		},
	})
}

func (s *Server) refreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refreshToken")

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}

	err = authToken.VerifyToken(cookie)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}

	userJwt, err := authToken.GetUserData(cookie)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	user := database.User{Username: userJwt.Username, Email: userJwt.Email, Uid: userJwt.Uid}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": authToken.CreateAccessToken(user),
	})

}

func (s *Server) getUserData(c *gin.Context) {
	jwt := c.GetHeader("Authorization")
	user := authToken.JwtUser{}

	if len(jwt) != 0 {
		jwt = strings.Split(jwt, " ")[1]

		err := authToken.VerifyToken(jwt)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
			return
		}

		user, err = authToken.GetUserData(jwt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
			fmt.Println(err)
			return
		}
	}

	userId := c.Param("uid")

	userData, err := s.db.GetUser(userId)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	if userId == user.Uid {
		c.JSON(http.StatusOK, gin.H{
			"user": gin.H{
				"Username":  userData.Username,
				"uid":       userData.Uid,
				"email":     userData.Email,
				"chatrooms": userData.Chatrooms,
			},
		})
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
