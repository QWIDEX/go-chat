package authToken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-chat/internal/database"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"

	"github.com/golang-jwt/jwt/v5"
)

// type of database.User that can be read using jwt token
type jwtUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Uid      string `json:"uid"`
}

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

func decodeBase64(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)

	if len(encoded)%4 != 0 {
		encoded += strings.Repeat("=", 4-len(encoded)%4)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return []byte{}, err
	}
	return decodedBytes, nil
}

func GetUserData(tokenString string) (jwtUser, error) {
	sect := strings.Split(tokenString, ".")

	decodedBytes, err := decodeBase64(sect[1])

	if err != nil {
		return jwtUser{}, err
	}

	userData := jwtUser{}

	err = json.Unmarshal(decodedBytes, &userData)

	if err != nil {
		return jwtUser{}, err
	}

	return userData, nil
}
