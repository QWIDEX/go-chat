package helpers

import (
	"fmt"
	authToken "go-chat/internal/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthHelper(c *gin.Context) (jwt string, user authToken.JwtUser, ok bool) {
	jwt, user, ok = "", authToken.JwtUser{}, false

	jwt = c.GetHeader("Authorization")

	if len(strings.Split(jwt, " ")) <= 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}

	jwt = strings.Split(jwt, " ")[1]

	err := authToken.VerifyToken(jwt)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "failed to authentificate"})
		return
	}

	user, err = authToken.GetUserData(jwt)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "something went wrong"})
		return
	}

	return jwt, user, true
}
