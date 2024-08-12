package authToken

import (
	"fmt"
	"go-chat/internal/database"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(user database.User) string {
	authSecret := os.Getenv("authSecret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "auth-server",
		"sub":      user.Username,
		"username": user.Username,
		"email":    user.Email,
		"uid":      user.Uid,
	})
	signedToken, _ := token.SignedString([]byte(authSecret))

	jwt.WithValidMethods([]string{"HS256"})

	return signedToken
}

func VerifyToken(tokenString string) error {
	authSecret := os.Getenv("authSecret")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(authSecret), nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
